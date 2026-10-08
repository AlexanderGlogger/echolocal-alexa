package echod

import (
	"context"
	"fmt"
	"io"
	"math"

	"github.com/zserge/microwakeword"

	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/lib/wake"
)

type variant struct {
	name string
	gain float64

	sum bool

	det  *microwakeword.Detector
	best float64
	hits int
}

func compareMixes(
	ctx context.Context,
	out io.Writer,
	source *mic.Source,
	m wake.Model,
	frames int,
) error {
	variants := []*variant{
		{name: "center", gain: 1},
		{name: "center x4", gain: 4},
		{name: "mean of 7", gain: 1, sum: true},
		{name: "mean of 7 x4", gain: 4, sum: true},
	}
	for _, v := range variants {
		det, err := microwakeword.NewDetector(m.Config)
		if err != nil {
			return err
		}
		v.det = det
	}

	raw, unlisten := source.ListenRaw()
	defer unlisten()

	fmt.Fprintf(
		out,
		"scoring %d mixes on the same audio, cutoff %.2f — say the wake word a few times\n\n",
		len(variants),
		m.Config.ProbabilityCutoff,
	)

	for range frames {
		select {
		case <-ctx.Done():
			return nil
		case frame, ok := <-raw:
			if !ok {
				return nil
			}
			chans := mic.Decode(frame)

			for _, v := range variants {
				samples := chans[mic.CenterMic]
				if v.sum {
					samples = meanOf(chans)
				}
				if v.gain != 1 {
					samples = amplify(samples, v.gain)
				}

				if v.det.ProcessAudio(samples) {
					v.hits++
				}
				if s := v.det.SlidingAverage(); s > v.best {
					v.best = s
				}
			}
		}
	}

	fmt.Fprintf(out, "%-14s %8s %6s\n", "mix", "best", "hits")
	for _, v := range variants {
		fmt.Fprintf(out, "%-14s %8.3f %6d\n", v.name, v.best, v.hits)
	}
	return nil
}

// Arrival delays across the array are under two samples at 16 kHz.
func meanOf(chans [][]int16) []int16 {
	out := make([]int16, len(chans[0]))
	for i := range out {
		var sum int
		for c := range chans {
			sum += int(chans[c][i])
		}
		out[i] = int16(math.Round(float64(sum) / float64(len(chans))))
	}
	return out
}
