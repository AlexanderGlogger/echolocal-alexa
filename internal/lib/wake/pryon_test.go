package wake

import (
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
	if got := Default([]string{sharedwake.DefaultModel, PryonID}); got != PryonID {
		t.Errorf("with Pryon installed the default is %q, want %q", got, PryonID)
	}
	if got := Default([]string{sharedwake.DefaultModel}); got != sharedwake.DefaultModel {
		t.Errorf("without Pryon the default is %q, want %q", got, sharedwake.DefaultModel)
	}
}
