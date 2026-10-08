package board

const (
	ColorBlack   = "black"
	ColorWhite   = "white"
	ColorUnknown = "unknown"
)

func (b Board) Color(idme func(string) string) string {
	if b.color == nil {
		return ColorUnknown
	}
	return b.color(idme)
}
