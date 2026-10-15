package sound

import (
	"encoding/binary"
	"testing"
)

// stereoFrame builds one interleaved 16-bit stereo frame.
func stereoFrame(l, r int16) []byte {
	b := make([]byte, bytesPerFrame)
	binary.LittleEndian.PutUint16(b[0:], uint16(l))
	binary.LittleEndian.PutUint16(b[2:], uint16(r))
	return b
}

func inputFrames(vals []int16) []byte {
	var out []byte
	for _, v := range vals {
		out = append(out, stereoFrame(v, v)...)
	}
	return out
}

func frameAt(b []byte, i int) int16 {
	return int16(binary.LittleEndian.Uint16(b[i*bytesPerFrame:]))
}

func TestResamplePCMUnchangedAtUnity(t *testing.T) {
	in := inputFrames([]int16{0, 100, 200, 300})
	out := resamplePCM(in, 1)
	if len(out) != len(in) {
		t.Fatalf("pitch 1 changed length: got %d want %d", len(out), len(in))
	}
	// pitch 1 returns the same backing slice unchanged.
	for i := 0; i < len(in); i++ {
		if out[i] != in[i] {
			t.Fatalf("pitch 1 modified byte %d", i)
		}
	}
}

func TestResamplePCMPitchUpShortens(t *testing.T) {
	in := inputFrames([]int16{0, 100, 200, 300, 400, 500, 600, 700})
	out := resamplePCM(in, 2)
	if want := len(in) / 2; len(out) != want {
		t.Fatalf("pitch 2 length: got %d want %d", len(out), want)
	}
	// Doubling pitch samples every other frame: 0, 200, 400, 600.
	for i, want := range []int16{0, 200, 400, 600} {
		if got := frameAt(out, i); got != want {
			t.Errorf("frame %d: got %d want %d", i, got, want)
		}
	}
}

func TestResamplePCMPitchDownLengthens(t *testing.T) {
	in := inputFrames([]int16{0, 300})
	out := resamplePCM(in, 0.5)
	if want := len(in) * 2; len(out) != want {
		t.Fatalf("pitch 0.5 length: got %d want %d", len(out), want)
	}
	// Linear interpolation halfway between 0 and 300 is 150.
	if got := frameAt(out, 1); got != 150 {
		t.Errorf("interpolated frame: got %d want 150", got)
	}
}

func TestResamplePCMNonPositivePitchIsUnity(t *testing.T) {
	in := inputFrames([]int16{0, 100, 200, 300})
	out := resamplePCM(in, 0)
	if len(out) != len(in) {
		t.Fatalf("pitch 0 should behave like 1: got %d want %d", len(out), len(in))
	}
}
