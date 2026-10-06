package config

// Screen is the panel, on a board that has one.
type Screen struct {
	// Theme is the palette by name, from internal/ui/theme.
	Theme string `json:"theme"`

	Keyboard KeyboardSize `json:"keyboard"`

	// Drawer is the side of the picture the rail of icons is pulled in from.
	Drawer Edge `json:"drawer"`

	// Volume is the side the volume card comes up on, left or right.
	Volume Edge `json:"volume"`

	// Marks is whether a corner of the panel says the microphones are cut or the camera covered.
	Marks bool `json:"marks"`

	Hours HourFormat `json:"hours"`
	Logo  bool       `json:"logo"`

	// Backlight is the panel's level in manual, and what the automatic curve is shifted by in auto.
	Backlight int        `json:"backlight"`
	Mode      ScreenMode `json:"mode"`

	Style string `json:"style"`
	Size  string `json:"size"`

	// Language is the text the panel shows, as a language tag. Empty is English.
	Language string `json:"language"`
}

const (
	UISizeMini    = "mini"
	UISizeCompact = "compact"
	UISizeMedium  = "medium"
	UISizeLarge   = "large"
	UISizeXLarge  = "xlarge"
)

func ScreenSizes() []string {
	return []string{UISizeMini, UISizeCompact, UISizeMedium, UISizeLarge, UISizeXLarge}
}

func (w ScreenWriter) Style(v string) error {
	return w.st.Update(func(c *Config) { c.Screen.Style = v })
}

func (w ScreenWriter) Size(v string) error {
	return w.st.Update(func(c *Config) { c.Screen.Size = v })
}

const (
	DefaultBacklight  = 70
	DefaultScreenMode = ModeAuto
)

func defaultScreen() Screen {
	return Screen{
		Theme:     DefaultTheme,
		Keyboard:  KeyboardCompact,
		Drawer:    DefaultEdge,
		Volume:    DefaultEdge,
		Marks:     true,
		Hours:     TwentyFourHour,
		Logo:      true,
		Backlight: DefaultBacklight,
		Mode:      DefaultScreenMode,
	}
}

// ScreenMode is how the brightness is decided.
type ScreenMode string

const (
	// ModeAuto sets the brightness from the ambient light sensor.
	ModeAuto ScreenMode = "auto"

	// ModeManual holds whatever the backlight is set to.
	ModeManual ScreenMode = "manual"
)

func (m ScreenMode) Label() string {
	switch m {
	case ModeAuto:
		return say.T("mode.auto")
	case ModeManual:
		return say.T("mode.manual")
	}
	return string(m)
}

type HourFormat string

const (
	TwentyFourHour HourFormat = "24"
	TwelveHour     HourFormat = "12"
)

func (h HourFormat) Label() string {
	switch h {
	case TwentyFourHour:
		return say.T("hours.24")
	case TwelveHour:
		return say.T("hours.12")
	}
	return string(h)
}

func HourFormats() []HourFormat { return []HourFormat{TwentyFourHour, TwelveHour} }

// DefaultTheme is what a device nobody has chosen for shows. It is also where an unreadable
// settings file lands, so it is the one worth being a sensible sight rather than a statement.
const DefaultTheme = "Midnight"

type ScreenWriter struct{ st *Store }

// All writes the panel's settings together, for a caller holding a whole one it has already changed.
func (w ScreenWriter) All(v Screen) error {
	return w.st.Update(func(c *Config) { c.Screen = v })
}

func (w ScreenWriter) Theme(v string) error {
	return w.st.Update(func(c *Config) { c.Screen.Theme = v })
}

func (w ScreenWriter) Drawer(v Edge) error {
	return w.st.Update(func(c *Config) { c.Screen.Drawer = v })
}

func (w ScreenWriter) Volume(v Edge) error {
	return w.st.Update(func(c *Config) { c.Screen.Volume = v })
}

func (w ScreenWriter) Marks(v bool) error {
	return w.st.Update(func(c *Config) { c.Screen.Marks = v })
}

func (w ScreenWriter) Hours(v HourFormat) error {
	return w.st.Update(func(c *Config) { c.Screen.Hours = v })
}

func (w ScreenWriter) Language(v string) error {
	return w.st.Update(func(c *Config) { c.Screen.Language = v })
}

func (w ScreenWriter) Keyboard(v KeyboardSize) error {
	return w.st.Update(func(c *Config) { c.Screen.Keyboard = v })
}

func (w ScreenWriter) Logo(v bool) error {
	return w.st.Update(func(c *Config) { c.Screen.Logo = v })
}

func (w ScreenWriter) Backlight(v int) error {
	return w.st.Update(func(c *Config) { c.Screen.Backlight = v })
}

func (w ScreenWriter) Mode(v ScreenMode) error {
	return w.st.Update(func(c *Config) { c.Screen.Mode = v })
}

type KeyboardSize string

const (
	KeyboardCompact  KeyboardSize = "compact"
	KeyboardStandard KeyboardSize = "standard"
	KeyboardLarge    KeyboardSize = "large"
)

func (k KeyboardSize) Label() string {
	switch k {
	case KeyboardCompact:
		return say.T("keyboard.compact")
	case KeyboardStandard:
		return say.T("keyboard.standard")
	case KeyboardLarge:
		return say.T("keyboard.large")
	}
	return string(k)
}

func KeyboardSizes() []KeyboardSize {
	return []KeyboardSize{KeyboardCompact, KeyboardStandard, KeyboardLarge}
}
