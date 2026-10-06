//go:build board_checkers || board_cronos

package settings

import (
	"slices"
	"strings"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/chromecast"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/lib/say"
	"github.com/ygelfand/echolocal/internal/ui/widget"
	"github.com/ygelfand/libcountertop/pkg/media/cast/protocols/youtube"
)

func init() {
	features = append(features,
		feature{
			label: func() string { return say.T("features.cast") },
			hint:  func() string { return say.T("features.cast.hint") },
			page:  castPage,
		},
	)
}

func castPage() *shell.Page {
	return &shell.Page{
		Title: say.T("features.cast"),
		Build: func() ([]widget.Row, []func(int)) {
			on := config.Get().Cast.Receiver
			return []widget.Row{
					{Label: say.T("cast.enable"), Kind: widget.Toggle, On: on},
					{Label: say.T("cast.apps"), Kind: widget.Chevron},
				}, []func(int){
					func(int) { chromecast.Get().SetReceiver(!on) },
					open(castAppsPage()),
				}
		},
	}
}

func castAppsPage() *shell.Page {
	return &shell.Page{
		Title: say.T("cast.apps"),
		Build: func() ([]widget.Row, []func(int)) {
			return []widget.Row{
					{Label: say.T("youtube.title"), Kind: widget.Chevron},
					{Label: say.T("cast.prime"), Kind: widget.Chevron},
				}, []func(int){
					open(youtubePage()),
					open(primePage()),
				}
		},
	}
}

func primePage() *shell.Page {
	return &shell.Page{
		Title: say.T("cast.prime"),
		Build: func() ([]widget.Row, []func(int)) {
			persist, skip := config.Get().Cast.Prime.Persist, config.Get().Cast.Prime.SkipIntro
			state := say.T("cast.prime.unregistered")
			if chromecast.Get().PrimeRegistered() {
				state = say.T("cast.prime.registered")
			}
			return []widget.Row{
					{
						Label: say.T("cast.prime.persist"),
						Hint:  say.T("cast.prime.persist.hint"),
						Kind:  widget.Toggle,
						On:    persist,
					},
					{Label: say.T("cast.prime.reset"), Kind: widget.Plain, Value: state},
					{
						Label: say.T("cast.prime.skipintro"),
						Hint:  say.T("cast.prime.skipintro.hint"),
						Kind:  widget.Toggle,
						On:    skip,
					},
				}, []func(int){
					func(int) { chromecast.Get().SetPrimePersist(!persist) },
					func(int) { chromecast.Get().ResetPrime() },
					func(int) { chromecast.Get().SetPrimeSkipIntro(!skip) },
				}
		},
	}
}

func youtubePage() *shell.Page {
	return &shell.Page{
		Title: say.T("youtube.title"),
		Build: func() ([]widget.Row, []func(int)) {
			c := config.Get().Cast.YouTube
			return []widget.Row{
					{
						Label: say.T("youtube.demand"),
						Hint:  say.T("youtube.demand.hint"),
						Kind:  widget.Toggle,
						On:    c.OnDemand,
					},
					{Label: say.T("youtube.skip"), Kind: widget.Chevron, Value: skipSays(c.Skip)},
					{
						Label: say.T("youtube.delay"),
						Hint:  say.T("youtube.delay.hint"),
						Kind:  widget.Slider,
						Level: delayLevel(
							c.LiveDelay,
						),
						Snap:  delaySnap,
						Value: delaySays(c.LiveDelay),
					},
				}, []func(int){
					func(int) { chromecast.Get().SetLoungeOnDemand(!c.OnDemand) },
					open(skipPage()),
					func(level int) { chromecast.Get().SetLiveDelay(delayOf(level)) },
				}
		},
	}
}

func skipPage() *shell.Page {
	return &shell.Page{
		Title: say.T("youtube.skip"),
		Build: func() ([]widget.Row, []func(int)) {
			skip := config.Get().Cast.YouTube.Skip
			rows := make([]widget.Row, 0, len(youtube.Categories))
			acts := make([]func(int), 0, len(youtube.Categories))
			for _, cat := range youtube.Categories {
				on := slices.Contains(skip, cat)
				rows = append(
					rows,
					widget.Row{Label: say.T("skip." + cat), Kind: widget.Toggle, On: on},
				)
				acts = append(
					acts,
					func(int) { chromecast.Get().SetSkip(strings.Join(toggled(skip, cat, !on), ",")) },
				)
			}
			return rows, acts
		},
	}
}

func toggled(skip []string, cat string, on bool) []string {
	out := slices.DeleteFunc(slices.Clone(skip), func(s string) bool { return s == cat })
	if on {
		out = append(out, cat)
	}
	return out
}

func skipSays(skip []string) string {
	if len(skip) == 0 {
		return say.T("youtube.skip.none")
	}
	return say.F("youtube.skip.some", map[string]any{"N": len(skip)})
}

const delaySpan = config.LiveDelayMost - config.LiveDelayLeast

func delayOf(level int) int {
	return config.LiveDelayLeast + (min(max(level, 0), 100)*delaySpan+50)/100
}

func delayLevel(seconds int) int {
	return (min(max(seconds-config.LiveDelayLeast, 0), delaySpan)*100 + delaySpan/2) / delaySpan
}

func delaySnap(level int) int { return delayLevel(delayOf(level)) }

func delaySays(
	seconds int,
) string {
	return say.F("youtube.delay.value", map[string]any{"N": seconds})
}
