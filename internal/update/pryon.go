package update

import (
	"errors"
	"os"

	"github.com/ygelfand/echolocal/internal/layout"
)

// errPryon refuses a self-update on a device installed with the Pryon companion. The channels serve
// upstream's builds, which have no Android media bridge: one would open the capture device directly
// while Pryon holds it, and the Alexa wake word would be gone.
var errPryon = errors.New("update: this device runs the Pryon/Alexa build; " +
	"update it with `echolocal pryon install --echod` instead")

func refusePryon() error {
	if st, err := os.Stat(layout.AndroidMediaJar); err == nil && st.Mode().IsRegular() {
		return errPryon
	}
	return nil
}
