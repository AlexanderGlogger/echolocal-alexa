package led

import "math"

var HomeAssistant = Color{R: 0x18, G: 0xBC, B: 0xF2}

func Solid(c Color) []Color {
	out := make([]Color, Segments)
	for i := range out {
		out[i] = c
	}
	return out
}

func Arc(fraction float64, c Color) []Color { return arcOf(fraction, Palette{c}) }

func Volume(fraction float64) []Color { return arcOf(fraction, vu) }

func arcOf(fraction float64, p Palette) []Color {
	fraction = math.Max(0, math.Min(1, fraction))
	lit := fraction * Segments

	out := make([]Color, Segments)
	for i := range Segments {
		remaining := lit - float64(i)
		if remaining <= 0 {
			break
		}
		part := math.Min(1, remaining)
		out[(11+i)%Segments] = scale(p.Along(float64(i)/(Segments-1)), part)
	}
	return out
}

func scale(c Color, f float64) Color {
	clamp := func(v byte) byte { return byte(math.Round(math.Max(0, math.Min(255, float64(v)*f)))) }
	return Color{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B)}
}

func add(a, b Color) Color {
	sum := func(x, y byte) byte {
		if int(x)+int(y) > 255 {
			return 255
		}
		return x + y
	}
	return Color{R: sum(a.R, b.R), G: sum(a.G, b.G), B: sum(a.B, b.B)}
}

func blend(a, b Color, f float64) Color {
	f = math.Max(0, math.Min(1, f))
	mix := func(x, y byte) byte {
		return byte(math.Round(float64(x) + (float64(y)-float64(x))*f))
	}
	return Color{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B)}
}

func hue(h float64) Color {
	x := math.Mod(h*6, 6)
	ramp := func(v float64) byte { return byte(math.Round(math.Max(0, math.Min(1, v)) * 255)) }
	switch int(x) {
	case 0:
		return Color{R: 255, G: ramp(x), B: 0}
	case 1:
		return Color{R: ramp(2 - x), G: 255, B: 0}
	case 2:
		return Color{R: 0, G: 255, B: ramp(x - 2)}
	case 3:
		return Color{R: 0, G: ramp(4 - x), B: 255}
	case 4:
		return Color{R: ramp(x - 4), G: 0, B: 255}
	default:
		return Color{R: 255, G: 0, B: ramp(6 - x)}
	}
}
