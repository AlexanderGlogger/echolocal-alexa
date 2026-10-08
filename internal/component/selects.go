package component

import (
	"log/slog"
	"slices"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/echolocal/internal/config"
)

func Bind[T config.Labelled](sel *esphome.Select, values []T, apply func(T) T, save func(T) error) {
	sel.Options = config.Labels(values)

	sel.OnCommand = func(label string) {
		want, ok := config.ByLabel(values, label)
		if !ok {
			slog.Warn("unknown option", "setting", sel.ObjectID, "value", label)
			return
		}

		settled := apply(want)
		sel.Set(settled.Label())
		if err := save(settled); err != nil {
			slog.Error("saving a setting failed", "setting", sel.ObjectID, "err", err)
		}
		slog.Info("setting changed", "setting", sel.ObjectID, "using", settled)
	}
}

func Restore[T config.Labelled](sel *esphome.Select, saved T, apply func(T) T) {
	settled := apply(saved)
	sel.Set(settled.Label())
	slog.Info("restored", "what", sel.ObjectID, "using", settled, "asked", saved)
}

const EffectNone = "None"

func SettleEffect(sel *esphome.Select, name string) string {
	if name == "" {
		name = EffectNone
	}
	if !slices.Contains(sel.Options, name) {
		slog.Warn(
			"no such effect, showing nothing instead",
			"setting",
			sel.ObjectID,
			"effect",
			name,
		)
		name = EffectNone
	}

	sel.Set(name)
	if name == EffectNone {
		return ""
	}
	return name
}

func BindEffect(
	sel *esphome.Select,
	choices []string,
	apply func(string),
	save func(string) error,
) {
	sel.Options = append([]string{EffectNone}, choices...)

	sel.OnCommand = func(chosen string) {
		settled := SettleEffect(sel, chosen)
		if apply != nil {
			apply(settled)
		}
		if err := save(settled); err != nil {
			slog.Error("saving a setting failed", "setting", sel.ObjectID, "err", err)
		}
		slog.Info("setting changed", "setting", sel.ObjectID, "using", sel.Get())
	}
}

func RestoreEffect(sel *esphome.Select, saved string, apply func(string), save func(string) error) {
	settled := SettleEffect(sel, saved)
	if apply != nil {
		apply(settled)
	}
	if settled != saved {
		if err := save(settled); err != nil {
			slog.Error("saving a corrected setting failed", "setting", sel.ObjectID, "err", err)
		}
	}
	slog.Info("restored", "what", sel.ObjectID, "using", sel.Get())
}

const EffectDefault = "Default"

func BindOverride(sel *esphome.Select, choices []string, save func(string) error) {
	sel.Options = append([]string{EffectDefault, EffectNone}, choices...)

	sel.OnCommand = func(chosen string) {
		settled := SettleOverride(sel, chosen)
		if err := save(settled); err != nil {
			slog.Error("saving a setting failed", "setting", sel.ObjectID, "err", err)
		}
		slog.Info("setting changed", "setting", sel.ObjectID, "using", sel.Get())
	}
}

func SettleOverride(sel *esphome.Select, name string) string {
	if name == "" {
		name = EffectDefault
	}
	if !slices.Contains(sel.Options, name) {
		slog.Warn(
			"no such effect, leaving it at the default",
			"setting",
			sel.ObjectID,
			"effect",
			name,
		)
		name = EffectDefault
	}

	sel.Set(name)
	if name == EffectDefault {
		return ""
	}
	return name
}

func RestoreOverride(sel *esphome.Select, saved string, save func(string) error) {
	settled := SettleOverride(sel, saved)
	if settled != saved {
		if err := save(settled); err != nil {
			slog.Error("saving a corrected setting failed", "setting", sel.ObjectID, "err", err)
		}
	}
	slog.Info("restored", "what", sel.ObjectID, "using", sel.Get())
}

func ChosenEffect(sel *esphome.Select) string {
	if name := sel.Get(); name != EffectNone {
		return name
	}
	return ""
}
