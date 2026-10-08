package led

import "time"

func roomMeter(p Palette, r Room) Frame {
	return func(time.Duration) []Color {
		return arcOf(level(r), p)
	}
}
