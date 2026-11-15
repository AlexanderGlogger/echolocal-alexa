package wake

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sync"

	esphome "github.com/ygelfand/go-esphome-device"

	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"

	"github.com/ygelfand/echolocal/internal/layout"
)

// KindPryon is Amazon's detector. Upstream scores it in a helper of its own, from a library and
// models it downloads; the companion's word shares the kind but has nothing for that engine to load.
const KindPryon = sharedwake.KindPryon

// PryonID is also the id upstream gives its own en-US Alexa. The Dots' saved choices hold it, so it
// stays, and the two are never offered together.
const PryonID = "pryon_alexa"

// PryonModel is the virtual model Home Assistant selects when the device-side companion is
// installed. It deliberately carries no path or inference configuration.
func PryonModel() sharedwake.Model {
	return sharedwake.Model{
		ID:        PryonID,
		Phrase:    "Alexa",
		Languages: []string{"en"},
		Kind:      KindPryon,
	}
}

// External reports whether m is the companion's word: the privileged Android companion owns its
// model and microphone, and sends only a wake event to echod. Upstream's own Pryon words carry the
// model they are scored from.
func External(m sharedwake.Model) bool { return m.Kind == KindPryon && m.Path == "" }

// companion reports whether the installer has recorded the companion's uid, which it does only after
// Android has scanned the package.
var companion = sync.OnceValue(func() bool {
	_, err := os.Stat(layout.PryonUIDPath)
	return err == nil
})

// native is upstream's own Pryon catalog, offered only where the companion is not installed. Both
// name en-US Alexa PryonID, and a saved choice of it has to keep meaning the detector that device
// was set up with.
func native(catalog sharedwake.Provider) sharedwake.Provider {
	if companion() {
		return absent{}
	}
	return catalog
}

type absent struct{}

func (absent) Models() []sharedwake.Model { return nil }

func (absent) Ensure(_ context.Context, id string, _ func(float32)) error {
	return fmt.Errorf("wake: %q is the companion's here", id)
}

// Library is the shared library plus Pryon, which is installed outside its model directory. Every
// caller reaches the library through Lib, so this is the one place Pryon has to be added.
type Library struct {
	*sharedwake.Library
	pryon bool
}

// withPryon offers Pryon where the companion is installed. The line it logs is what
// `echolocal pryon install` waits for.
func withPryon(l *sharedwake.Library) *Library {
	if companion() {
		slog.Info("wake word offered", "id", PryonID, "engine", KindPryon)
	}
	return &Library{Library: l, pryon: companion()}
}

func (l *Library) Ours() []sharedwake.Model { return l.with(l.Library.Ours()) }

func (l *Library) Ensure(ids []string) []sharedwake.Model { return l.with(l.Library.Ensure(ids)) }

// Answer advertises Pryon as one of the device's own words, so an offer of the same phrase from Home
// Assistant is shadowed by it the way it would be by an installed model.
func (l *Library) Answer(offered []esphome.ExternalWakeWord) []esphome.WakeWord {
	if !l.pryon {
		return l.Library.Answer(offered)
	}
	p := PryonModel()
	offered = slices.DeleteFunc(slices.Clone(offered),
		func(o esphome.ExternalWakeWord) bool { return o.Phrase == p.Phrase })
	return append(l.Library.Answer(offered),
		esphome.WakeWord{ID: p.ID, Phrase: p.Phrase, TrainedLanguages: p.Languages})
}

func (l *Library) with(models []sharedwake.Model) []sharedwake.Model {
	if l.pryon {
		models = append(models, PryonModel())
	}
	return models
}

// Default is the wake word a device with no saved choice starts on. An installer that set the
// companion up chose Alexa on purpose; everywhere else it is upstream's default.
func Default(installed []sharedwake.Model) string {
	if slices.ContainsFunc(installed, External) {
		return PryonID
	}
	return sharedwake.DefaultModel
}
