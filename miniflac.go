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

func Decode(flacData []byte) (*Waveform, error) {
	if len(flacData) == 0 {
		return nil, errors.New("miniflac: input data is empty")
	}
	var decoder C.miniflac_t
	C.miniflac_init(&decoder, C.MINIFLAC_CONTAINER_UNKNOWN)
	var p runtime.Pinner
	defer p.Unpin()
	p.Pin(unsafe.SliceData(flacData))
	samples := make([][]C.int32_t, maxChannels)
	ptrs := make([]*C.int32_t, maxChannels)
	for i := range samples {
		samples[i] = make([]C.int32_t, maxSamplesPerFrame)
		ptr := unsafe.SliceData(samples[i])
		p.Pin(ptr)
		ptrs[i] = ptr
	}
	var (
		used  C.uint32_t
		shift C.uint8_t
	)
	var waveform *Waveform
	// TODO: can we get the number of total frames in flacData here?
	for {
		res := C.miniflac_decode(
			&decoder,
			(*C.uint8_t)(unsafe.SliceData(flacData)),
			C.uint32_t(len(flacData)),
			&used,
			(**C.int32_t)(unsafe.SliceData(ptrs)),
		)
		if res != C.MINIFLAC_OK {
			break
		}
		if waveform == nil {
			// TODO: can we use the first header's channel, bps, and sample rate?
			// TODO: is this okay? can we just use the first frame's header info as global one?
			waveform = &Waveform{
				Channels:       int(decoder.frame.header.channels),
				SampleRate:     int(decoder.frame.header.sample_rate),
				SourceBitDepth: int(decoder.frame.header.bps),
			}
		}
		flacData = flacData[used:]
		// TODO: does bps change from frame to frame?
		switch bps := decoder.frame.header.bps; {
		case bps <= 8:
			shift = 8 - bps
		case bps <= 16:
			shift = 16 - bps
		case bps <= 24:
			shift = 24 - bps
		case bps <= 32:
			shift = 32 - bps
		default:
			return nil, errors.New("miniflac: invalid bps")
		}
		for i := range decoder.frame.header.block_size {
			for j := range decoder.frame.header.channels {
				// TODO: is this bitshift operation required?
				waveform.Samples = append(waveform.Samples, int(C.uint32_t(samples[j][i])<<shift))
			}
		}
	}
	if waveform == nil {
		return nil, errors.New("miniflac: no frames decoded")
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
	Samples        []int
	SourceBitDepth int
}
