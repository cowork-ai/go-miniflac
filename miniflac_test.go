package miniflac

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"os"
	"testing"
)

var writeGolden = flag.Bool("write-golden", false, "")

func TestDecodeNoData(t *testing.T) {
	_, err := Decode(nil)
	if got, want := err, ErrNoData; !errors.Is(got, want) {
		t.Errorf("Decode=%v, want=%v", got, want)
	}
	_, err = Decode([]byte{})
	if got, want := err, ErrNoData; !errors.Is(got, want) {
		t.Errorf("Decode=%v, want=%v", got, want)
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			"96kHz24bit",
			"./testdata/Sample_BeeMoved_96kHz24bit.flac",
			"./testdata/Sample_BeeMoved_96kHz24bit.pcm",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := os.ReadFile(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			wav, err := Decode(bs)
			if err != nil {
				t.Fatal(err)
			}
			if *writeGolden {
				f, err := os.Create(tt.out)
				if err != nil {
					t.Fatal(err)
				}
				if err := writeS32LE(f, wav); err != nil {
					t.Fatal(err)
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			}
			f, err := os.CreateTemp(t.TempDir(), "test_decode_*.pcm")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())
			if err := writeS32LE(f, wav); err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			if got, want := fileHash(t, f.Name()), fileHash(t, tt.out); got != want {
				t.Errorf("fileHash=%v, want=%v", got, want)
			}
		})
	}
}

func writeS32LE(f io.Writer, w *Waveform) error {
	const maxBitDepth = 32
	if w.SourceBitDepth < 4 || w.SourceBitDepth > maxBitDepth {
		return errors.New("invalid bit depth")
	}
	writer := bufio.NewWriter(f)
	shift := maxBitDepth - w.SourceBitDepth
	for _, s := range w.Samples {
		// Sample is shifted to align the most significant bits for 32-bit output.
		sample32 := int32(uint32(s) << shift)
		if err := binary.Write(writer, binary.LittleEndian, sample32); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func fileHash(t *testing.T, filename string) string {
	t.Helper()
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
