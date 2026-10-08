package audio

import (
	"fmt"
	"io"
	"math"
)

const FullScale = 1 << 23

func DecodeS24LE3(b []byte) int32 {
	v := int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16
	if v&0x800000 != 0 {
		v |= ^0xFFFFFF
	}
	return v
}

type Stats struct {
	N        int64
	Peak     int32
	sumSq    float64
	sumSqErr float64
	nonzero  int64
}

func (s *Stats) Add(v int32) {
	s.N++
	if a := v; a < 0 {
		if -a > s.Peak {
			s.Peak = -a
		}
	} else if a > s.Peak {
		s.Peak = a
	}
	s.sumSq += float64(v) * float64(v)

	resid := v - (v>>8)<<8
	s.sumSqErr += float64(resid) * float64(resid)
	if resid != 0 {
		s.nonzero++
	}
}

func (s *Stats) RMS() float64 {
	if s.N == 0 {
		return 0
	}
	return math.Sqrt(s.sumSq / float64(s.N))
}

func (s *Stats) TruncErrRMS() float64 {
	if s.N == 0 {
		return 0
	}
	return math.Sqrt(s.sumSqErr / float64(s.N))
}

// Acoustic capture sets the low 8 bits; a digital loopback in the top 16 bits never does.
func (s *Stats) LowBitsUsed() float64 {
	if s.N == 0 {
		return 0
	}
	return float64(s.nonzero) / float64(s.N)
}

func DBFS(v float64) float64 {
	if v <= 0 {
		return math.Inf(-1)
	}
	return 20 * math.Log10(v/FullScale)
}

func WriteReport(w io.Writer, st []Stats) {
	fmt.Fprintf(w, "\n%3s  %10s  %10s  %12s  %10s  %s\n",
		"ch", "peak dBFS", "rms dBFS", "trunc-err dB", "trunc SNR", "low-8-bits")
	for c := range st {
		s := &st[c]
		if s.N == 0 {
			continue
		}
		rms, errRMS := s.RMS(), s.TruncErrRMS()

		snr := math.Inf(1)
		if errRMS > 0 {
			snr = 20 * math.Log10(rms/errRMS)
		}

		note := ""
		switch {
		case s.Peak == 0:
			note = "  <- silent"
		case s.LowBitsUsed() == 0:
			note = "  <- digital loopback, not a mic"
		}

		fmt.Fprintf(w, "%3d  %10.1f  %10.1f  %12.1f  %8.1f dB  %9.1f%%%s\n",
			c, DBFS(float64(s.Peak)), DBFS(rms), DBFS(errRMS), snr,
			100*s.LowBitsUsed(), note)
	}
}
