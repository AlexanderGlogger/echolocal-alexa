package setup

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/android/dns"
	"github.com/ygelfand/echolocal/internal/android/firewall"
	"github.com/ygelfand/echolocal/internal/android/prop"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/layout"
)

const pinControl = "/sys/devices/soc/1000b000.pinctrl/mt_gpio"

type Action struct {
	Name   string
	Reason string
	Do     func() error
}

func stop(service, reason string) Action {
	return Action{
		Name:   "stop " + service,
		Reason: reason,
		Do:     func() error { return prop.Stop(service) },
	}
}

var Actions = []Action{
	stop("meshmgrservice", "the other half of the BLE mesh stack"),
	stop("whad_cc", "whole-home audio's control channel, left with nothing once whad is hidden"),
	stop("vitals_service", "collects the device's vitals for Amazon"),
	stop("perfmonitord", "Amazon's performance monitoring"),
	stop("perfrecoveryd", "Amazon's performance monitoring"),
	stop("avahi-daemon", "Amazon's mDNS, for Spotify Connect; echod advertises itself"),
	stop("drm", "drmserver, for protected media playback"),
	{
		Name: "take the pin controller back from mediaserver",
		Reason: "the vendor audio HAL clears the microphone mute line while it sets up an input path, " +
			"which unmutes a device the user left muted",
		// Amazon's csm_audio_init.sh chowns the mute LED gpio to root.media.
		Do: func() error {
			if component.Board().MuteDir != "" {
				return nil
			}
			return os.Chmod(pinControl, 0o644)
		},
	},
	{
		Name:   "silence AmazonUsageStatsService",
		Reason: "logs the network state on a timer from inside system_server",
		// The persist form of this property is 39 bytes, past what Android 5.1 accepts.
		Do: func() error { return prop.Set("log.tag.AmazonUsageStatsService", "S") },
	},
	{
		Name:   "silence the video encoder",
		Reason: "MediaTek's encoder logs every frame it encodes",
		Do:     func() error { return prop.Set("log.tag.VDO_LOG", "S") },
	},
	{
		Name:   "quiet the image processor",
		Reason: "MediaTek's MDP warns on every camera frame that it could not raise its thread priority",
		Do:     func() error { return prop.Set("log.tag.MDP", "E") },
	},
	{
		Name:   "size the runtime to the cores that are present",
		Reason: "the kernel hotplugs them, so Go sees however many were online when it started",
		Do:     procs,
	},
	{
		Name:   "point the resolver at the nameservers the platform knows",
		Reason: "there is no /etc/resolv.conf, so Go's resolver has nowhere to look",
		Do:     func() error { dns.Use(); return nil },
	},
	{
		Name:   "point the verifier at the platform's root certificates",
		Reason: "crypto/x509 reads the Android store only in a GOOS=android build, and this one is linux",
		// SSL_CERT_DIR replaces the directories crypto/x509 scans, never the files it reads.
		Do: func() error { return os.Setenv("SSL_CERT_DIR", certDirs) },
	},
	{
		Name:   "give temporary files a directory",
		Reason: "there is no /tmp, and go-glyph writes its fonts there",
		Do:     temp,
	},
}

func temp() error {
	if err := os.RemoveAll(layout.TempDir); err != nil {
		return err
	}
	if err := os.MkdirAll(layout.TempDir, 0o700); err != nil {
		return err
	}
	return os.Setenv("TMPDIR", layout.TempDir)
}

const certDirs = "/system/etc/security/cacerts:/data/misc/keychain/certs-added"

// A stop sent before the boot completes is undone by the start that follows it.
var Late = []Action{
	stop(
		"shblemeshd",
		"BLE mesh daemon left with nothing to talk to once its service package is hidden",
	),
	stop("sntpd", "Amazon's time client; echod keeps the clock"),
	stop("ace_messaging", "Amazon's push channel and the device's to-do queue"),
	{
		Name: "put the firewall jumps back",
		Reason: "the boot completing is what starts firewall.sh, which flushes INPUT: our chains keep " +
			"their rules but nothing reaches them again until INPUT jumps to them",
		Do: firewall.Ensure,
	},
}

const (
	bootProp = "sys.boot_completed"
	bootWait = 3 * time.Minute
	bootPoll = 5 * time.Second

	// wifisvc powers the radio and starts the supplicant; started again, it resets wifi.
	wifiService = "wifisvc"
	wifiDriver  = "wlan.driver.status"
	wifiLink    = "/sys/class/net/wlan0/operstate"
)

func Apply() { apply("boot setup", Actions) }

type Setup struct{}

func init() {
	component.Register(
		sharedcomponent.Hardware,
		func() Setup { return Setup{} },
		sharedcomponent.Order(1),
	)
}

func (Setup) Name() string { return "platform setup" }

func (Setup) Start(context.Context) error {
	Apply()
	return nil
}

func (Setup) Run(ctx context.Context) error {
	ApplyLate(ctx)
	return nil
}

func ApplyLate(ctx context.Context) {
	if component.Board().SignalsBoot && !waitFor(ctx, "the boot to finish", bootCompleted) {
		return
	}
	apply("late setup", Late)
	releaseWifi(ctx)
}

func releaseWifi(ctx context.Context) {
	if state, err := prop.Get("init.svc." + wifiService); err != nil || state != "running" {
		return
	}
	if !waitFor(ctx, "wifi before stopping "+wifiService, wifiUp) {
		return
	}
	if err := prop.Stop(wifiService); err != nil {
		slog.Error("late setup failed", "action", "stop "+wifiService, "err", err)
		return
	}
	slog.Info("stopped once wifi was up", "service", wifiService)
}

func apply(what string, actions []Action) {
	var failed int
	for _, a := range actions {
		if err := a.Do(); err != nil {
			slog.Error(what+" failed", "action", a.Name, "reason", a.Reason, "err", err)
			failed++
		}
	}
	slog.Info(what+" applied", "actions", len(actions), "failed", failed)
}

func bootCompleted() bool {
	v, err := prop.Get(bootProp)
	return err == nil && v == "1"
}

func wifiUp() bool {
	driver, err := prop.Get(wifiDriver)
	if err != nil || driver != "ok" {
		return false
	}
	link, err := os.ReadFile(wifiLink)
	return err == nil && strings.TrimSpace(string(link)) == "up"
}

func waitFor(ctx context.Context, what string, ready func() bool) bool {
	tick := time.NewTicker(bootPoll)
	defer tick.Stop()
	giveUp := time.After(bootWait)

	for {
		if ready() {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-giveUp:
			slog.Warn("gave up waiting for "+what, "after", bootWait)
			return false
		case <-tick.C:
		}
	}
}
