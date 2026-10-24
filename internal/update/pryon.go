package update

import (
	"errors"
	"os"
	"strings"

	"github.com/ygelfand/echolocal/internal/layout"
)

// forkRepo publishes the Pryon/Alexa builds. channel.go and the Makefile's REPO both point at it, so a
// device fetches from here and the manifests CI writes name binaries here.
const forkRepo = "https://github.com/AlexanderGlogger/echolocal-alexa"

// errPryon refuses a self-update on a device installed with the Pryon companion unless it comes from the
// fork. Upstream's builds have no Android media bridge: one would open the capture device directly while
// Pryon holds it, and the Alexa wake word would be gone.
var errPryon = errors.New(
	"update: this device runs the Pryon/Alexa build and only installs releases from " +
		forkRepo + "; update it with `echolocal pryon install --echod` instead",
)

func refusePryon(m Manifest) error { return refusePryonAt(layout.AndroidMediaJar, m) }

// refusePryonAt is refusePryon for a bridge at jar, which in practice means a test.
func refusePryonAt(jar string, m Manifest) error {
	if st, err := os.Stat(jar); err != nil || !st.Mode().IsRegular() {
		return nil
	}
	if fromFork(m) {
		return nil
	}
	return errPryon
}

// fromFork holds when every binary m names is one of the fork's release downloads.
func fromFork(m Manifest) bool {
	var urls []string
	if m.URL != "" {
		urls = append(urls, m.URL)
	}
	for _, b := range m.Binaries {
		urls = append(urls, b.URL)
	}
	for _, builds := range m.Boards {
		for _, b := range builds {
			urls = append(urls, b.URL)
		}
	}

	for _, u := range urls {
		if !strings.HasPrefix(u, forkRepo+"/releases/download/") {
			return false
		}
	}
	return len(urls) > 0
}
