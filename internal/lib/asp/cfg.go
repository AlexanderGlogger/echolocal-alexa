package asp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Coefficient files hold one float per line with a trailing comma.
func readFloats(path string, want int) ([]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make([]float32, 0, want)
	s := bufio.NewScanner(f)
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		text = strings.TrimSuffix(text, ",")
		if text == "" || strings.HasPrefix(text, "/*") || strings.HasPrefix(text, "//") {
			continue
		}

		v, err := strconv.ParseFloat(text, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %q is not a coefficient", path, line, text)
		}
		out = append(out, float32(v))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(out) != want {
		return nil, fmt.Errorf("%s: %d coefficients, expected %d", path, len(out), want)
	}
	return out, nil
}

type mbcl struct {
	Bypass          bool       `json:"Bypass"`
	PreFilterBypass bool       `json:"PreFilterBypass"`
	InVol           float64    `json:"inVol"`
	NumBands        int        `json:"NumBands"`
	Crossovers      []float64  `json:"FilterBank FC"`
	Bands           []mbclBand `json:"Bands Definition"`
	Full            mbclLimit  `json:"Full-band limiter"`
}

// Thresholds are dB below full scale and release is in milliseconds; the file has no attack times.
type mbclBand struct {
	CompInVol   float64 `json:"comp_inVol"`
	CompRatio   float64 `json:"comp_ratio"`
	CompThresh  float64 `json:"comp_thresh"`
	CompGainMin float64 `json:"comp_gainMin"`
	LimInVol    float64 `json:"lim_inVol"`
	LimThresh   float64 `json:"lim_thresh"`
	LimRelease  float64 `json:"lim_release"`
}

type mbclLimit struct {
	LimInVol   float64 `json:"lim_inVol"`
	LimThresh  float64 `json:"lim_thresh"`
	LimRelease float64 `json:"lim_release"`
}

// The vendor file is JSON with C comments, which the vendor loader strips.
func readMBCL(path string) (mbcl, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return mbcl{}, err
	}

	var m mbcl
	if err := json.Unmarshal(stripComments(raw), &m); err != nil {
		return mbcl{}, fmt.Errorf("%s: %w", path, err)
	}

	if n := len(m.Bands); n != m.NumBands || n != len(m.Crossovers)+1 {
		return mbcl{}, fmt.Errorf("%s: %d band definitions, NumBands %d, %d crossovers",
			path, n, m.NumBands, len(m.Crossovers))
	}
	for i, fc := range m.Crossovers {
		if fc <= 0 || fc >= Rate/2 {
			return mbcl{}, fmt.Errorf("%s: crossover %d at %g Hz is not in band", path, i, fc)
		}
		if i > 0 && fc <= m.Crossovers[i-1] {
			return mbcl{}, fmt.Errorf("%s: crossovers are not ascending: %v", path, m.Crossovers)
		}
	}
	return m, nil
}

func stripComments(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); {
		switch {
		case b[i] == '"':
			j := i + 1
			for j < len(b) && b[j] != '"' {
				if b[j] == '\\' {
					j++
				}
				j++
			}
			j = min(j+1, len(b))
			out = append(out, b[i:j]...)
			i = j

		case b[i] == '/' && i+1 < len(b) && b[i+1] == '/':
			for i < len(b) && b[i] != '\n' {
				i++
			}

		case b[i] == '/' && i+1 < len(b) && b[i+1] == '*':
			i += 2
			for i+1 < len(b) && (b[i] != '*' || b[i+1] != '/') {
				i++
			}
			i = min(i+2, len(b))

		default:
			out = append(out, b[i])
			i++
		}
	}
	return out
}
