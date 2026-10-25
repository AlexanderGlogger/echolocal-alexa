package speaker

import (
	"math"
	"testing"

	"github.com/ygelfand/echolocal/internal/config"
)

func TestRouteForEachModeAndPlug(t *testing.T) {
	auto, on, off := config.OutputAuto, config.OutputOn, config.OutputOff
	for _, c := range []struct {
		plugged            bool
		speaker, headphone config.OutputMode
		want               route
	}{
		{false, auto, auto, route{out: OutputSpeaker, speaker: true}},
		{true, auto, auto, route{out: OutputHeadphone, speaker: false}},
		{true, on, auto, route{out: OutputHeadphone, speaker: true}},
		{true, on, off, route{out: OutputSpeaker, speaker: true}},
		{false, off, on, route{out: OutputHeadphone, speaker: false}},
		{true, off, off, route{out: OutputSpeaker, speaker: false}},
		{false, auto, off, route{out: OutputSpeaker, speaker: true}},
	} {
		if got := routeFor(c.plugged, c.speaker, c.headphone); got != c.want {
			t.Errorf("plugged=%t speaker=%s headphone=%s: got %+v, want %+v",
				c.plugged, c.speaker, c.headphone, got, c.want)
		}
	}
}

func TestGainForStep(t *testing.T) {
	cases := []struct {
		out  Output
		step int
		want float32
	}{
		{OutputSpeaker, 0, 0},
		{OutputSpeaker, 30, 1},
		{OutputSpeaker, 15, 0.398},
		{OutputSpeaker, 5, 0.0708},
		{OutputHeadphone, 0, 0},
		{OutputHeadphone, 30, 1},
		{OutputHeadphone, 15, 0.112},
	}

	for _, c := range cases {
		got := gainForStep(c.out, c.step)
		if math.Abs(float64(got-c.want)) > 0.001 {
			t.Errorf("gainForStep(%s, %d) = %v, want %v", c.out, c.step, got, c.want)
		}
	}
}

func TestGainForStepClamps(t *testing.T) {
	if got := gainForStep(OutputSpeaker, -5); got != 0 {
		t.Errorf("gainForStep(speaker, -5) = %v, want 0", got)
	}
	if got := gainForStep(OutputSpeaker, 99); got != 1 {
		t.Errorf("gainForStep(speaker, 99) = %v, want 1", got)
	}
}

func TestGainForStepUnknownOutput(t *testing.T) {
	if got := gainForStep(Output("bluetooth"), 30); got != 1 {
		t.Errorf("gainForStep(bluetooth, 30) = %v, want 1", got)
	}
}
