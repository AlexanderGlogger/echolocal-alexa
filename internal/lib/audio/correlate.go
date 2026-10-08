package audio

import (
	"fmt"
	"io"
	"math"
)

func Channels(raw []byte, chans int) [][]float64 {
	frameBytes := chans * 3
	frames := len(raw) / frameBytes

	out := make([][]float64, chans)
	for c := range out {
		out[c] = make([]float64, frames)
	}
	for f := 0; f < frames; f++ {
		off := f * frameBytes
		for c := 0; c < chans; c++ {
			out[c][f] = float64(DecodeS24LE3(raw[off+c*3:]))
		}
	}
	return out
}

func Correlation(a, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}

	var ma, mb float64
	for i := 0; i < n; i++ {
		ma += a[i]
		mb += b[i]
	}
	ma /= float64(n)
	mb /= float64(n)

	var num, da, db float64
	for i := 0; i < n; i++ {
		x, y := a[i]-ma, b[i]-mb
		num += x * y
		da += x * x
		db += y * y
	}
	if da == 0 || db == 0 {
		return 0
	}
	return num / math.Sqrt(da*db)
}

func CorrelationMatrix(w io.Writer, ch [][]float64) {
	n := len(ch)

	fmt.Fprintf(w, "\n%4s", "")
	for c := 0; c < n; c++ {
		fmt.Fprintf(w, "%7d", c)
	}
	fmt.Fprintln(w)

	for i := 0; i < n; i++ {
		fmt.Fprintf(w, "%4d", i)
		for j := 0; j < n; j++ {
			fmt.Fprintf(w, "%7.3f", Correlation(ch[i], ch[j]))
		}
		fmt.Fprintln(w)
	}
}

func MeanAbsCorrelation(ch [][]float64, i int) float64 {
	if len(ch) < 2 {
		return 0
	}
	var sum float64
	for j := range ch {
		if j == i {
			continue
		}
		sum += math.Abs(Correlation(ch[i], ch[j]))
	}
	return sum / float64(len(ch)-1)
}

type Topology struct {
	Center       int
	CenterSpread float64
	Spreads      []float64
	ByDistance   [4]float64
}

func AnalyzeTopology(ch [][]float64, perimeter []int, center int) Topology {
	t := Topology{Center: center}

	spread := func(i int) float64 {
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, j := range perimeter {
			if j == i {
				continue
			}
			v := Correlation(ch[i], ch[j])
			lo, hi = math.Min(lo, v), math.Max(hi, v)
		}
		return hi - lo
	}

	t.CenterSpread = spread(center)
	for _, i := range perimeter {
		t.Spreads = append(t.Spreads, spread(i))
	}

	n := len(perimeter)
	var sums [4]float64
	var counts [4]int
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			d := b - a
			if d > n/2 {
				d = n - d
			}
			sums[d] += Correlation(ch[perimeter[a]], ch[perimeter[b]])
			counts[d]++
		}
	}
	for d := 1; d <= 3; d++ {
		if counts[d] > 0 {
			t.ByDistance[d] = sums[d] / float64(counts[d])
		}
	}
	return t
}

func (t Topology) Monotonic() bool {
	return t.ByDistance[1] > t.ByDistance[2] && t.ByDistance[2] > t.ByDistance[3]
}

// SpeedOfSound in air, m/s, near room temperature.
const SpeedOfSound = 343.0

func LoudestWindow(ref []float64, n int) int {
	if n <= 0 || n >= len(ref) {
		return 0
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += ref[i] * ref[i]
	}
	best, bestSum := 0, sum
	for i := n; i < len(ref); i++ {
		sum += ref[i]*ref[i] - ref[i-n]*ref[i-n]
		if sum > bestSum {
			best, bestSum = i-n+1, sum
		}
	}
	return best
}

func SliceAll(ch [][]float64, start, n int) [][]float64 {
	out := make([][]float64, len(ch))
	for i, c := range ch {
		end := start + n
		if end > len(c) {
			end = len(c)
		}
		if start >= end {
			out[i] = c
			continue
		}
		out[i] = c[start:end]
	}
	return out
}

// 36 mm of separation is about 105 us, 1.7 samples at 16 kHz.
func SubSampleLag(a, b []float64, maxLag int) (float64, float64) {
	corr := make([]float64, 2*maxLag+1)
	for i := range corr {
		lag := i - maxLag
		var x, y []float64
		switch {
		case lag < 0:
			x, y = a[-lag:], b[:len(b)+lag]
		case lag > 0:
			x, y = a[:len(a)-lag], b[lag:]
		default:
			x, y = a, b
		}
		corr[i] = Correlation(x, y)
	}

	peak := 0
	for i, v := range corr {
		if v > corr[peak] {
			peak = i
		}
	}
	lag := float64(peak - maxLag)

	if peak > 0 && peak < len(corr)-1 {
		l, c, r := corr[peak-1], corr[peak], corr[peak+1]
		if d := l - 2*c + r; d != 0 {
			lag += 0.5 * (l - r) / d
		}
	}
	return lag, corr[peak]
}

func ArrivalOffset(a, b []float64, maxLag int) (float64, float64) {
	lag, corr := SubSampleLag(a, b, maxLag)
	return -lag, corr
}

func FitBearing(offsets []float64) (bearingDeg, amplitude float64) {
	n := len(offsets)
	if n == 0 {
		return 0, 0
	}
	var sc, ss float64
	for i, off := range offsets {
		a := 2 * math.Pi * float64(i) / float64(n)
		sc += off * math.Cos(a)
		ss += off * math.Sin(a)
	}
	// Sum of cos^2 over evenly spaced angles is n/2, and the model carries a minus sign.
	cx, cy := -sc/(float64(n)/2), -ss/(float64(n)/2)

	deg := math.Atan2(cy, cx) * 180 / math.Pi
	if deg < 0 {
		deg += 360
	}
	return deg, math.Hypot(cx, cy)
}

func InPlaneLag(radiusMM float64, rate int) float64 {
	return radiusMM / 1000 / SpeedOfSound * float64(rate)
}

func LagToDistance(lagSamples float64, rate int) float64 {
	return lagSamples / float64(rate) * SpeedOfSound * 1000
}

func BestLag(a, b []float64, maxLag int) (int, float64) {
	best, bestVal := 0, math.Inf(-1)
	for lag := -maxLag; lag <= maxLag; lag++ {
		var x, y []float64
		switch {
		case lag < 0:
			x, y = a[-lag:], b[:len(b)+lag]
		case lag > 0:
			x, y = a[:len(a)-lag], b[lag:]
		default:
			x, y = a, b
		}
		if v := Correlation(x, y); v > bestVal {
			best, bestVal = lag, v
		}
	}
	return best, bestVal
}
