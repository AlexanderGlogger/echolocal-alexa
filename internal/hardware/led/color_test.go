package led

import "testing"

func lit(frame []Color) (count int, last Color) {
	for i := range Segments {
		c := frame[(11+i)%Segments]
		if c == (Color{}) {
			break
		}
		count, last = count+1, c
	}
	return count, last
}

func TestArcGrowsBySegmentAndDimsTheLeadingOne(t *testing.T) {
	white := Color{R: 0xFF, G: 0xFF, B: 0xFF}

	for _, tc := range []struct {
		fraction float64
		want     int
	}{{0, 0}, {0.5, 6}, {1, Segments}} {
		if got, _ := lit(Arc(tc.fraction, white)); got != tc.want {
			t.Errorf("Arc(%v) lit %d segments, want %d", tc.fraction, got, tc.want)
		}
	}

	_, edge := lit(Arc(5.5/Segments, white))
	if edge.R == 0 || edge.R == 0xFF {
		t.Errorf("the leading segment is %v, want it dimmed but not out", edge)
	}
}

func TestVolumeIsGreenLowAndRedHigh(t *testing.T) {
	_, quiet := lit(Volume(0.25))
	if quiet.G <= quiet.R {
		t.Errorf("a quarter volume leads with %v, want green", quiet)
	}

	_, loud := lit(Volume(1))
	if loud.R <= loud.G {
		t.Errorf("full volume leads with %v, want red", loud)
	}

	full := Volume(1)
	for _, fraction := range []float64{0.25, 0.5, 0.75} {
		frame := Volume(fraction)
		n, _ := lit(frame)
		if n < 2 {
			t.Fatalf("Volume(%v) lit %d segments", fraction, n)
		}
		at := (11 + n - 2) % Segments
		if frame[at] != full[at] {
			t.Errorf("segment %d is %v at %v and %v when full", at, frame[at], fraction, full[at])
		}
	}
}
