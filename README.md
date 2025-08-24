# go-miniflac

[![Go Reference](https://pkg.go.dev/badge/github.com/cowork-ai/go-miniflac.svg)](https://pkg.go.dev/github.com/cowork-ai/go-miniflac)

[go-miniflac](https://github.com/cowork-ai/go-miniflac) is a Go binding for the [miniflac](https://github.com/jprjr/miniflac) C library. The following is the miniflac
description from its author, [@jprjr](https://github.com/jprjr).

> A single-file C library for decoding FLAC streams. Does not use any C library functions, does not allocate any memory.

go-miniflac has a very simple interface, one function and one struct, and has zero external dependencies. However, Cgo
must be enabled to compile this package.

## Interface

```go
// Decode parses a FLAC byte slice and returns a [Waveform].
func Decode(flacData []byte) (*Waveform, error)

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
```

## Examples

### How to convert an flac file to a WAV file using [go-audio/wav](https://github.com/go-audio/wav)

Check out [examples/flac-to-wav](https://github.com/cowork-ai/go-miniflac/blob/main/examples/flac-to-wav/main.go)

## Taskfile.yml

Many useful commands are in two `Taskfile.yml` files: [Taskfile.yml](https://github.com/cowork-ai/go-miniflac/blob/main/Taskfile.yml) and [examples/Taskfile.yml](https://github.com/cowork-ai/go-miniflac/blob/main/examples/Taskfile.yml). To run the tasks, you need to install [go-task/task](https://github.com/go-task/task), which works similarly to [GNU Make](https://www.gnu.org/software/make/).

## Dockerfile

Check out the [Dockerfile](https://github.com/cowork-ai/go-miniflac/blob/main/Dockerfile) for an example of using `golang:1.25-bookworm` and `gcr.io/distroless/base-debian12` to run `go-miniflac` with Cgo enabled.

```bash
docker build -t cowork-ai/go-miniflac .
cat ./testdata/Sample_BeeMoved_96kHz24bit.flac | docker run --rm -i cowork-ai/go-miniflac | ffplay -autoexit -i pipe:
```

Note that the `gcr.io/distroless/static-debian12` image does not work because it lacks `glibc`.
