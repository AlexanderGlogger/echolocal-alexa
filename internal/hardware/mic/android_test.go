package mic

import (
	"encoding/binary"
	"testing"
)

func TestReframerCutsWholeFramesAcrossReads(t *testing.T) {
	stream := make([]byte, FrameSamples*2*3)
	for i := range len(stream) / 2 {
		binary.LittleEndian.PutUint16(stream[i*2:], uint16(i))
	}

	var got [][]int16
	var r reframer
	// Uneven reads, including an odd byte count that splits a sample.
	for _, n := range []int{1, 639, 3, 700, 577} {
		r.push(stream[:n], func(f []int16) { got = append(got, f) })
		stream = stream[n:]
	}

	if len(got) != 3 {
		t.Fatalf("got %d frames, want 3", len(got))
	}
	for f, frame := range got {
		if len(frame) != FrameSamples {
			t.Fatalf("frame %d has %d samples, want %d", f, len(frame), FrameSamples)
		}
		for i, v := range frame {
			if want := int16(f*FrameSamples + i); v != want {
				t.Fatalf("frame %d sample %d = %d, want %d", f, i, v, want)
			}
		}
	}
	if len(r.pending) != 0 {
		t.Fatalf("%d bytes left over, want 0", len(r.pending))
	}
}
