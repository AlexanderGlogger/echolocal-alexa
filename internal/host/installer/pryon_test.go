package installer

import (
	"reflect"
	"testing"
)

func TestUniqueLocales(t *testing.T) {
	got := unique([]string{"en-GB", "en", "en-GB", "", "en-US"})
	want := []string{"en-GB", "en", "en-US"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine("\n/a/first\n/a/second\n"); got != "/a/first" {
		t.Fatalf("got %q", got)
	}
}

func TestPryonUserIDPattern(t *testing.T) {
	match := userIDPattern.FindStringSubmatch("Packages:\n  userId=32065\n")
	if len(match) != 2 || match[1] != "32065" {
		t.Fatalf("got %v", match)
	}
}

func TestPackagePath(t *testing.T) {
	tests := map[string]struct {
		output string
		want   string
	}{
		"visible": {
			output: "package:/system/priv-app/SpeechInteractionManager/SpeechInteractionManager.apk\n",
			want:   "/system/priv-app/SpeechInteractionManager/SpeechInteractionManager.apk",
		},
		"hidden": {
			output: "package:/system/priv-app/SpeechInteractionManager/SpeechInteractionManager.apk=amazon.speech.sim\n",
			want:   "/system/priv-app/SpeechInteractionManager/SpeechInteractionManager.apk",
		},
		"wrong package": {
			output: "package:/system/other.apk=example.other\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := packagePath(test.output, "amazon.speech.sim"); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestPryonRuntimeReadyRequiresLiveAudioEndToEnd(t *testing.T) {
	poc := "PRYON_CAPTURE_ACTIVE\nPRYON_SHARED_PCM_FIRST_FRAME\nPRYON_READY"
	helper := "shared Pryon capture first frame bytes=512"
	echo := "id=pryon_alexa engine=pryon"
	if !pryonRuntimeReady(poc, helper, echo) {
		t.Fatal("complete live microphone handshake was not ready")
	}

	tests := map[string]struct {
		poc, helper, echo string
	}{
		"native detector not ready": {
			poc:    "PRYON_CAPTURE_ACTIVE\nPRYON_SHARED_PCM_FIRST_FRAME",
			helper: helper,
			echo:   echo,
		},
		"native recorder stalled": {
			poc:    "PRYON_SHARED_PCM_FIRST_FRAME\nPRYON_READY",
			helper: helper,
			echo:   echo,
		},
		"shared reader empty": {
			poc:    "PRYON_CAPTURE_ACTIVE\nPRYON_READY",
			helper: helper,
			echo:   echo,
		},
		"helper received no PCM": {poc: poc, echo: echo},
		"Alexa not advertised":   {poc: poc, helper: helper},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if pryonRuntimeReady(test.poc, test.helper, test.echo) {
				t.Fatal("incomplete handshake reported ready")
			}
		})
	}
}

// Both lines are as each generation of echod writes them to logcat.
func TestMDNSAdvertisedReadsEitherEchod(t *testing.T) {
	for line, want := range map[string]bool{
		"[   41.20] advertising what=esphome addrs=192.168.1.20":            true,
		"advertising over mdns name=kitchen port=6053 addrs=192.168.1.20":   true,
		"[   40.02] waiting for an address before advertising what=esphome": false,
		"[   41.20] advertising what=sendspin addrs=192.168.1.20":           false,
	} {
		if got := mdnsAdvertised(line); got != want {
			t.Errorf("mdnsAdvertised(%q) = %t, want %t", line, got, want)
		}
	}
}
