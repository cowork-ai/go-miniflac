# go-miniflac

## How to Test

First, install Go 1.25+ and Taskfile 3+. You can install them via Homebrew or by following the instructions on the official [Go](https://go.dev/doc/install) and [Taskfile](https://taskfile.dev/docs/installation) installation pages. Then, run the following command to convert a FLAC file to a WAV file.

```bash
task test/flac-to-wav
```

If you have FFmpeg installed, `ffplay` will automatically play the output WAV file after the conversion.
