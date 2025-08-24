package miniflac

// #define MINIFLAC_IMPLEMENTATION
// #define MINIFLAC_API
// #define MINIFLAC_PRIVATE static inline
// #include "miniflac.h"
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

const (
	maxChannels        = 8
	maxSamplesPerFrame = 65_535
)

var (
	ErrNoData             = errors.New("miniflac: no data provided")
	ErrStreamInfoNotFound = errors.New("miniflac: STREAMINFO not found")
)

// Decode parses a FLAC byte slice and returns a [Waveform].
func Decode(flacData []byte) (*Waveform, error) {
	// NOTE: unsafe.SliceData returns an invalid pointer if a slice is non-nil but cap() is zero, e.g., []byte{}.
	if len(flacData) == 0 {
		return nil, ErrNoData
	}
	var p runtime.Pinner
	defer p.Unpin()
	p.Pin(unsafe.SliceData(flacData))
	var decoder C.miniflac_t
	C.miniflac_init(&decoder, C.MINIFLAC_CONTAINER_UNKNOWN)
	var used C.uint32_t
	if res := C.miniflac_sync(
		&decoder, (*C.uint8_t)(unsafe.SliceData(flacData)), C.uint32_t(len(flacData)), &used,
	); res != C.MINIFLAC_OK {
		return nil, errors.New("miniflac: miniflac_sync failed")
	}
	flacData = flacData[used:]
	var (
		sampleRate      C.uint32_t
		channels        C.uint8_t
		bps             C.uint8_t
		totalSamples    C.uint64_t
		streamInfoFound bool
	)
	for C.miniflac_is_metadata(&decoder) == 1 {
		if C.miniflac_metadata_is_streaminfo(&decoder) == 1 {
			var res C.MINIFLAC_RESULT
			res = C.miniflac_streaminfo_sample_rate(
				&decoder, (*C.uint8_t)(unsafe.SliceData(flacData)), C.uint32_t(len(flacData)), &used, &sampleRate,
			)
			if res != C.MINIFLAC_OK {
				return nil, errors.New("miniflac: miniflac_streaminfo_sample_rate failed")
			}
			flacData = flacData[used:]
			res = C.miniflac_streaminfo_channels(
				&decoder, (*C.uint8_t)(unsafe.SliceData(flacData)), C.uint32_t(len(flacData)), &used, &channels,
			)
			if res != C.MINIFLAC_OK {
				return nil, errors.New("miniflac: miniflac_streaminfo_channels failed")
			}
			if channels == 0 || channels > maxChannels {
				return nil, errors.New("miniflac: invalid channels")
			}
			flacData = flacData[used:]
			res = C.miniflac_streaminfo_bps(
				&decoder, (*C.uint8_t)(unsafe.SliceData(flacData)), C.uint32_t(len(flacData)), &used, &bps,
			)
			if res != C.MINIFLAC_OK {
				return nil, errors.New("miniflac: miniflac_streaminfo_bps failed")
			}
			if bps < 4 || bps > 32 {
				return nil, errors.New("miniflac: invalid bps")
			}
			flacData = flacData[used:]
			res = C.miniflac_streaminfo_total_samples(
				&decoder, (*C.uint8_t)(unsafe.SliceData(flacData)), C.uint32_t(len(flacData)), &used, &totalSamples,
			)
			if res != C.MINIFLAC_OK {
				return nil, errors.New("miniflac: miniflac_streaminfo_total_samples failed")
			}
			flacData = flacData[used:]
			if totalSamples == 0 {
				return nil, errors.New("miniflac: invalid total samples")
			}
			streamInfoFound = true
			break
		}
		if res := C.miniflac_sync(
			&decoder, (*C.uint8_t)(unsafe.SliceData(flacData)), C.uint32_t(len(flacData)), &used,
		); res != C.MINIFLAC_OK {
			return nil, errors.New("miniflac: miniflac_sync failed")
		}
		flacData = flacData[used:]
	}
	if !streamInfoFound {
		return nil, ErrStreamInfoNotFound
	}
	waveform := &Waveform{
		Channels:       int(channels),
		SampleRate:     int(sampleRate),
		Samples:        make([]int, totalSamples*C.uint64_t(channels)),
		SourceBitDepth: int(bps),
	}
	samples := make([][]C.int32_t, maxChannels)
	ptrs := make([]*C.int32_t, maxChannels)
	for i := range samples {
		samples[i] = make([]C.int32_t, maxSamplesPerFrame)
		ptr := unsafe.SliceData(samples[i])
		p.Pin(ptr)
		ptrs[i] = ptr
	}
	var sampleCount int
	for {
		if res := C.miniflac_decode(
			&decoder,
			(*C.uint8_t)(unsafe.SliceData(flacData)),
			C.uint32_t(len(flacData)),
			&used,
			(**C.int32_t)(unsafe.SliceData(ptrs)),
		); res != C.MINIFLAC_OK {
			break
		}
		flacData = flacData[used:]
		header := decoder.frame.header
		for i := range header.block_size {
			for j := range header.channels {
				waveform.Samples[sampleCount] = int(samples[j][i])
				sampleCount++
			}
		}
	}
	if sampleCount != len(waveform.Samples) {
		return nil, errors.New("miniflac: sample count mismatch")
	}
	return waveform, nil
}

// Waveform represents decoded PCM audio data. 🎶
type Waveform struct {
	// Channels is the number of audio channels (e.g., 1 for mono, 2 for stereo).
	Channels int
	// SampleRate is the number of samples per second (e.g., 44100 Hz).
	SampleRate int
	// Samples contains the interleaved audio data.
	Samples []int
	// SourceBitDepth is the original bit depth of the audio.
	SourceBitDepth int
}
