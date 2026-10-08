package all

import (
	"slices"
	"strings"
	"testing"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/led"
)

var registered = []string{
	"ble_advertisements",
	"bluetooth_proxy",
	"button_action",
	"button_mute",
	"button_volume_down",
	"button_volume_up",
	"cached_data",
	"check_for_updates",
	"cpu_cores",
	"cpu_cores_online",
	"cpu_temperature",
	"device_chime",
	"failure_effect",
	"firmware",
	"follow_up_1",
	"follow_up_2",
	"free_space",
	"hardware_board",
	"hardware_color",
	"headphones",
	"ip_address",
	"keep_recordings_1",
	"keep_recordings_2",
	"last_heard",
	"last_reply",
	"last_wake_word",
	"load_average",
	"lux",
	"max_listen_1",
	"max_listen_2",
	"max_think_1",
	"max_think_2",
	"media_duck_level",
	"memory_available",
	"metrics_interval",
	"mic_mute",
	"microphone_cancel_echo",
	"microphone_gain",
	"microphone_leveling",
	"microphone_mixing",
	"microphone_sensitivity",
	"min_cores",
	"mute_led_brightness",
	"mute_sound",
	"noise_layer_1",
	"noise_layer_2",
	"purge_cache",
	"radio_temperature",
	"remote_adb",
	"reply_buffer_1",
	"reply_buffer_2",
	"reply_delivery_1",
	"reply_delivery_2",
	"replying_effect_1",
	"replying_effect_2",
	"restart",
	"ring",
	"ring_muted",
	"room_floor",
	"room_level",
	"room_reaction",
	"segment_1",
	"segment_10",
	"segment_11",
	"segment_12",
	"segment_2",
	"segment_3",
	"segment_4",
	"segment_5",
	"segment_6",
	"segment_7",
	"segment_8",
	"segment_9",
	"sendspin",
	"sendspin_artist",
	"sendspin_state",
	"sendspin_title",
	"speaker",
	"speaker_eq",
	"stop_word_sensitivity",
	"test_playback",
	"thinking_effect_1",
	"thinking_effect_2",
	"time_zone",
	"timers",
	"update_channel",
	"update_outcome",
	"update_status",
	"verify_certificates",
	"voice_resampling",
	"wake_assistant_1",
	"wake_assistant_2",
	"wake_effect_1",
	"wake_effect_2",
	"wake_threshold_1",
	"wake_threshold_2",
	"wake_tone_1",
	"wake_tone_2",
	"wifi_received",
	"wifi_sent",
	"wifi_signal",
}

func TestSupportedBoardsHaveUniqueEntityKeys(t *testing.T) {
	was := component.Board()
	t.Cleanup(func() { component.Default().Use(was) })
	for _, b := range []board.Board{board.Biscuit, board.Checkers, board.Cronos} {
		t.Run(b.Codename, func(t *testing.T) {
			component.Default().Use(b)
			entities := esphome.NewEntities()
			if err := entities.Add(component.Default().Entities()...); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEveryComponentStillRegisters(t *testing.T) {
	var got []string
	for _, e := range component.Default().Entities() {
		got = append(got, objectID(e))
	}
	slices.Sort(got)

	if !slices.Equal(got, registered) {
		t.Errorf("registered entities changed\n got: %s\nwant: %s",
			strings.Join(got, " "), strings.Join(registered, " "))
	}
}

func TestARinglessBoardHasNoRingEntities(t *testing.T) {
	ringless := board.Board{Codename: "ringless", Model: "A board with no ring"}

	component.Default().Use(ringless)
	t.Cleanup(func() { component.Default().Use(board.Biscuit) })

	var got []string
	for _, e := range component.Default().Entities() {
		got = append(got, objectID(e))
	}

	gone := []string{
		"ring", "segment_1", "segment_12",
		"ring_muted", "failure_effect", "room_reaction",
		"wake_effect_1", "thinking_effect_1", "replying_effect_1",
		"wake_effect_2", "thinking_effect_2", "replying_effect_2",
	}
	for _, id := range gone {
		if slices.Contains(got, id) {
			t.Errorf("%q is still registered on a board with no ring", id)
		}
	}

	for _, e := range component.Default().Entities() {
		sel, ok := e.(*esphome.Select)
		if !ok {
			continue
		}
		for _, opt := range sel.Options {
			if slices.Contains(led.EffectNames(), opt) {
				t.Errorf("%q still offers the ring animation %q", objectID(e), opt)
			}
		}
	}

	for _, id := range []string{"speaker", "mic_mute", "timers", "hardware_board", "firmware"} {
		if !slices.Contains(got, id) {
			t.Errorf("%q went missing on a board with no ring", id)
		}
	}
}

func TestNoDuplicateObjectIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range component.Default().Entities() {
		id := objectID(e)
		if seen[id] {
			t.Errorf("two components registered %q", id)
		}
		seen[id] = true
	}
}

func TestEveryComponentIsNamed(t *testing.T) {
	for _, c := range component.Default().All() {
		if c.Name() == "" {
			t.Errorf("%T has no name", c)
		}
	}
}

func objectID(e esphome.Entity) string {
	type based interface{ Object() string }
	if b, ok := e.(based); ok {
		return b.Object()
	}
	return ""
}
