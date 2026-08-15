package wake

import (
	"os"
	"slices"

	esphome "github.com/ygelfand/go-esphome-device"

	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"

	"github.com/ygelfand/echolocal/internal/layout"
)

// KindPryon is Amazon's firmware detector. It has no TFLite path: the privileged Android companion
// owns its model and microphone, and sends only a wake event to echod.
const KindPryon sharedwake.Kind = "pryon"

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

// Library is the shared library plus Pryon, which is installed outside its model directory. Every
// caller reaches the library through Lib, so this is the one place Pryon has to be added.
type Library struct {
	*sharedwake.Library
	pryon bool
}

// withPryon offers Pryon once the installer has recorded the companion's uid, which it does only
// after Android has scanned the package.
func withPryon(l *sharedwake.Library) *Library {
	_, err := os.Stat(layout.PryonUIDPath)
	return &Library{Library: l, pryon: err == nil}
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

// Default is the wake word a device with no saved choice starts on. An installer that set Pryon up
// chose Alexa on purpose; everywhere else it is upstream's default.
func Default(installed []string) string {
	if slices.Contains(installed, PryonID) {
		return PryonID
	}
	return sharedwake.DefaultModel
}
