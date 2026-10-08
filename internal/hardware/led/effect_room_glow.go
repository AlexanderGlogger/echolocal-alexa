package led

import "time"

// A lit segment whines audibly on this hardware.

func roomGlow(p Palette, r Room) Frame {
	return func(time.Duration) []Color {
		return around(p, level(r))
	}
}
