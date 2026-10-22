package component

import (
	"fmt"
	"slices"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/libcountertop/pkg/say"
	setting "github.com/ygelfand/libcountertop/pkg/settings"
)

const (
	RingGroup    setting.Group = "Ring"
	MuteGroup    setting.Group = "Microphone"
	SpeakerGroup setting.Group = "Speaker"
)

type Picked struct {
	Name   string
	Select *esphome.Select
	Switch *esphome.Switch
}

type Picks struct{}

func Rows(domain string, group setting.Group, picks ...Picked) *setting.Controls[Picks] {
	byName := map[string]Picked{}
	rows := make([]setting.Setting[Picks], 0, len(picks))
	for _, p := range picks {
		byName[p.Name] = p
		rows = append(rows, row(domain, group, p))
	}
	return &setting.Controls[Picks]{
		Table: setting.NewTable(domain, []setting.Group{group}, rows,
			setting.Messages{Text: say.T, Missing: say.Missing}),
		Read: func() Picks { return Picks{} },
		Save: func(s setting.Setting[Picks], v string) error {
			var at Picks
			if err := s.Write(&at, v); err != nil {
				return err
			}
			p := byName[s.Name]
			if p.Switch != nil {
				on, _ := setting.Boolean(v)
				p.Switch.OnCommand(on)
				return nil
			}
			p.Select.OnCommand(v)
			return nil
		},
	}
}

func row(domain string, group setting.Group, p Picked) setting.Setting[Picks] {
	s := setting.Setting[Picks]{Name: p.Name, Group: group, NoEntity: true}
	if sw := p.Switch; sw != nil {
		s.Kind = setting.Toggle
		s.Read = func(*Picks) string { return setting.OnOff(sw.Get()) }
		s.Write = func(_ *Picks, v string) error {
			if _, ok := setting.Boolean(v); !ok {
				return fmt.Errorf("%s.%s takes on or off, not %q", domain, p.Name, v)
			}
			return nil
		}
		return s
	}
	sel := p.Select
	s.Kind = setting.Choice
	s.More = func() []setting.Option {
		out := make([]setting.Option, 0, len(sel.Options))
		for _, o := range sel.Options {
			out = append(out, setting.Option{Value: o, Label: o})
		}
		return out
	}
	s.Read = func(*Picks) string { return sel.Get() }
	s.Write = func(_ *Picks, v string) error {
		if !slices.Contains(sel.Options, v) {
			return fmt.Errorf("%s.%s: no choice called %q", domain, p.Name, v)
		}
		return nil
	}
	return s
}
