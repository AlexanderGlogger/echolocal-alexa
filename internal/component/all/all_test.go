package all

import (
	"slices"
	"testing"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/led"
)

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
		"ring_incoming_call", "ring_in_call",
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
