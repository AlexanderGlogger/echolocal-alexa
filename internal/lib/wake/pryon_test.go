package wake

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	esphome "github.com/ygelfand/go-esphome-device"

	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"
)

// installed writes a manifest for each id so the directory reads back as those models.
func installed(t *testing.T, phrases map[string]string, pryon bool) *Library {
	t.Helper()
	dir := t.TempDir()

	for id, phrase := range phrases {
		body := `{"wake_word":"` + phrase + `","model":"` + id + `.tflite"}`
		if err := os.WriteFile(
			filepath.Join(dir, id+".tflite"),
			[]byte("not a model"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, id+".json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &Library{Library: sharedwake.NewLibrary(dir), pryon: pryon}
}

func ids[T any](items []T, id func(T) string) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, id(it))
	}
	slices.Sort(out)
	return out
}

func modelID(m sharedwake.Model) string { return m.ID }
func wordID(w esphome.WakeWord) string  { return w.ID }

func TestPryonIsOneOfOursOnlyWhenInstalled(t *testing.T) {
	for _, pryon := range []bool{false, true} {
		l := installed(t, map[string]string{"okay_nabu": "Okay Nabu"}, pryon)

		want := []string{"okay_nabu"}
		if pryon {
			want = []string{"okay_nabu", PryonID}
		}
		if got := ids(l.Ours(), modelID); !slices.Equal(got, want) {
			t.Errorf("pryon=%v: ours %v, want %v", pryon, got, want)
		}
		if got := ids(l.Ensure([]string{PryonID}), modelID); !slices.Equal(got, want) {
			t.Errorf("pryon=%v: ensured %v, want %v", pryon, got, want)
		}
	}
}

// Home Assistant keys its selects by phrase, so an offered Alexa next to Pryon's would be two
// options with one name. Pryon shadows it the way an installed model would.
func TestPryonShadowsAnOfferedAlexa(t *testing.T) {
	l := installed(t, map[string]string{"okay_nabu": "Okay Nabu"}, true)

	words := l.Answer([]esphome.ExternalWakeWord{
		{ID: "alexa", Phrase: "Alexa", URL: "http://ha/alexa.json"},
		{ID: "hey_jarvis", Phrase: "Hey Jarvis", URL: "http://ha/hey_jarvis.json"},
	})
	if got, want := ids(
		words,
		wordID,
	), []string{
		"hey_jarvis",
		"okay_nabu",
		PryonID,
	}; !slices.Equal(
		got,
		want,
	) {
		t.Fatalf("advertised %v, want %v", got, want)
	}
}

func TestAnswerIsUpstreamsWithoutPryon(t *testing.T) {
	l := installed(t, map[string]string{"okay_nabu": "Okay Nabu"}, false)

	words := l.Answer(
		[]esphome.ExternalWakeWord{{ID: "alexa", Phrase: "Alexa", URL: "http://ha/alexa.json"}},
	)
	if got, want := ids(words, wordID), []string{"alexa", "okay_nabu"}; !slices.Equal(got, want) {
		t.Fatalf("advertised %v, want %v", got, want)
	}
}

func TestDefaultPrefersPryon(t *testing.T) {
	nabu := sharedwake.Model{ID: sharedwake.DefaultModel}
	if got := Default([]sharedwake.Model{nabu, PryonModel()}); got != PryonID {
		t.Errorf("with the companion installed the default is %q, want %q", got, PryonID)
	}
	if got := Default([]sharedwake.Model{nabu}); got != sharedwake.DefaultModel {
		t.Errorf("without Pryon the default is %q, want %q", got, sharedwake.DefaultModel)
	}
	if got := Default([]sharedwake.Model{nabu, upstreams()}); got != sharedwake.DefaultModel {
		t.Errorf("with upstream's Alexa the default is %q, want %q", got, sharedwake.DefaultModel)
	}
}

// upstreams is en-US Alexa as upstream's own Pryon catalog lists it: the same id and kind as the
// companion's, plus the model its engine loads.
func upstreams() sharedwake.Model {
	return sharedwake.Model{
		ID:      PryonID,
		Phrase:  "Pryon: Alexa",
		Kind:    sharedwake.KindPryon,
		Path:    "/data/misc/echolocal/models/pryon/models/en-US/pryon.manifest",
		Keyword: "ALEXA",
	}
}

type theirs struct{}

func (theirs) Models() []sharedwake.Model { return []sharedwake.Model{upstreams()} }

func (theirs) Ensure(context.Context, string, func(float32)) error { return nil }

func installs(t *testing.T, on bool) {
	t.Helper()
	was := companion
	companion = func() bool { return on }
	t.Cleanup(func() { companion = was })
}

// Both detectors call en-US Alexa pryon_alexa. A Dot's saved choice of it must keep meaning the
// detector the Dot was installed with, so only one of them is ever on offer.
func TestOnlyOnePryonAlexaIsOffered(t *testing.T) {
	for _, on := range []bool{false, true} {
		installs(t, on)
		l := withPryon(sharedwake.NewLibrary(t.TempDir(), native(theirs{})))

		ours := l.Ensure([]string{PryonID})
		if got := ids(ours, modelID); !slices.Equal(got, []string{PryonID}) {
			t.Fatalf("companion=%v: ours %v, want one %q", on, got, PryonID)
		}
		if got := External(ours[0]); got != on {
			t.Errorf("companion=%v: the offered Alexa is the companion's: %v", on, got)
		}
		words := l.Answer(nil)
		if len(words) != 1 || words[0].ID != PryonID {
			t.Errorf("companion=%v: advertised %v, want one %q", on, words, PryonID)
		}
	}
}
