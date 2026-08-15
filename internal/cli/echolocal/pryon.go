package echolocal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/ygelfand/libcountertop/pkg/host/adb"
	"github.com/ygelfand/libcountertop/pkg/host/prompt"

	"github.com/ygelfand/echolocal/internal/host/installer"
	"github.com/ygelfand/echolocal/internal/layout"
)

// Where the Android builds leave the companion and the bridge, relative to the source tree.
const (
	defaultPryonAPK = "android/pryon/build/EchoLocalPryon.apk"
	defaultHelper   = "android/amazon-helper/build/amazon-helper.jar"
)

func newPryonCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "pryon",
		Short: "Add Amazon's native Alexa wake word to an Echo Dot 2 on Fire OS 5",
		Long: "Installs the Pryon companion and the Android media bridge onto an Echo Dot 2 that already\n" +
			"runs echod on Fire OS 5, so Amazon's own on-device detector can be picked as the Alexa wake\n" +
			"word. Amazon's libraries and models stay on the device and are discovered there.",
		Args: cobra.NoArgs,
	}
	c.AddCommand(newPryonInstallCmd(), newPryonStatusCmd(), newPryonRemoveCmd())
	return c
}

func newPryonInstallCmd() *cobra.Command {
	var (
		serial   string
		apk      string
		helper   string
		echod    string
		doReboot bool
		noReboot bool
	)

	c := &cobra.Command{
		Use:   "install",
		Short: "Install, or update, the Pryon companion and the Android media bridge",
		Long: "Stages the companion and the bridge, reboots when Android has to scan the companion, then\n" +
			"configures and verifies Pryon. Safe to re-run; a run with nothing new skips the reboot.\n\n" +
			"--echod replaces the installed echod in the same run. Fire OS 5 runs the arm64 build:\n" +
			"make build-echod DOT_ARCH=arm64",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, out := cmd.Context(), cmd.OutOrStdout()

			var cfg installer.PryonConfig
			var err error
			if cfg.APK, err = readBuild(
				apk,
				"Pryon companion",
				"android/pryon/build.ps1",
			); err != nil {
				return err
			}
			if cfg.Helper, err = readBuild(
				helper,
				"Android media bridge",
				"android/amazon-helper/build.ps1",
			); err != nil {
				return err
			}
			if echod != "" {
				if cfg.Echod, err = os.ReadFile(echod); err != nil {
					return fmt.Errorf("reading echod: %w", err)
				}
			}

			d, err := connect(ctx, out, serial)
			if err != nil {
				return err
			}

			var needsReboot bool
			if err := render(
				ctx,
				out,
				"Installing Pryon",
				"✓ Pryon companion and Android media bridge staged",
				func(report installer.Reporter) error {
					var err error
					needsReboot, err = installer.PryonInstall(ctx, d, cfg, report)
					return err
				},
			); err != nil {
				return err
			}

			if needsReboot {
				if err := rebootForPryon(
					ctx,
					out,
					d,
					rebootChoiceOf(doReboot, noReboot),
				); err != nil {
					return err
				}
			}

			if err := render(
				ctx,
				out,
				"Finalizing Alexa",
				"✓ Pryon ready and Alexa advertised to Home Assistant",
				func(report installer.Reporter) error {
					return installer.FinalizePryon(ctx, d, report)
				},
			); err != nil {
				return err
			}
			fmt.Fprintf(out, "\n%s\n", styleDetail.Render(
				"Choose Alexa in the Wake word select of the Home Assistant assistant it should start.",
			))
			return nil
		},
	}

	c.Flags().StringVar(&serial, "serial", "", "device serial, when more than one is connected")
	c.Flags().StringVar(&apk, "apk", defaultPryonAPK, "Pryon companion APK to install")
	c.Flags().StringVar(&helper, "helper", defaultHelper, "Android media bridge JAR to install")
	c.Flags().
		StringVar(&echod, "echod", "", "echod binary to install in the same run (the arm64 build)")
	c.Flags().
		BoolVar(&doReboot, "reboot", false, "reboot without asking when Android has to scan the companion")
	c.Flags().BoolVar(&noReboot, "no-reboot", false, "stop before the reboot instead of asking")
	c.MarkFlagsMutuallyExclusive("reboot", "no-reboot")
	return c
}

// rebootForPryon has the boot happen that makes Android scan the companion. Finalizing cannot run
// before it, so declining ends the run with what to do next rather than half an install.
func rebootForPryon(ctx context.Context, out io.Writer, d *adb.Device, choice rebootChoice) error {
	again := errors.New("Pryon is staged but Android has not scanned the companion yet; " +
		"reboot the device and run `echolocal pryon install` again to finish")

	switch choice {
	case rebootNo:
		return again
	case rebootAsk:
		if !prompt.IsTerminal() {
			return fmt.Errorf("%w (or pass --reboot)", again)
		}
		yes, err := prompt.Confirm(
			ctx,
			out,
			"Reboot now? Android only scans the Pryon companion while it boots.",
		)
		if err != nil && !errors.Is(err, prompt.ErrCanceled) {
			return err
		}
		if !yes {
			return again
		}
	}

	if err := render(ctx, out, "Rebooting", "✓ device came back and echod started on its own",
		func(report installer.Reporter) error {
			return installer.RebootAndWait(ctx, d, report)
		}); err != nil {
		return err
	}

	scanned, err := installer.PryonScanned(d)
	if err != nil {
		return err
	}
	if !scanned {
		return fmt.Errorf("Android did not scan %s during the reboot", layout.PryonAPK)
	}
	return nil
}

func readBuild(path, what, script string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf(
			"%s not found at %s; build it with %s, or pass its path",
			what,
			path,
			script,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("reading the %s: %w", what, err)
	}
	return b, nil
}

func newPryonStatusCmd() *cobra.Command {
	var serial string

	c := &cobra.Command{
		Use:   "status",
		Short: "Show whether Pryon and the Android media bridge are installed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			d, err := connect(cmd.Context(), cmd.OutOrStdout(), serial)
			if err != nil {
				return err
			}
			s := installer.ReadPryonState(d)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, styleTitle.Render("Pryon status"))
			row := func(k, v string) { fmt.Fprintf(out, "  %s %s\n", styleKey.Render(k), v) }

			if s.APIListening {
				row(
					"ESPHome API",
					styleDone.Render("listening")+fmt.Sprintf(" on tcp/%d", layout.Port),
				)
			} else {
				row("ESPHome API", styleFail.Render("not listening"))
			}
			if s.AndroidMedia {
				row("audio", styleDone.Render("Android media bridge"))
			} else {
				row("audio", styleSkip.Render("direct ALSA"))
			}
			switch {
			case s.Installed && s.Configured:
				row("Alexa", styleDone.Render("Pryon installed and configured"))
			case s.Installed:
				row("Alexa", styleFail.Render("Pryon installed but not finalized"))
			default:
				row("Alexa", styleSkip.Render("not installed"))
			}
			return nil
		},
	}
	c.Flags().StringVar(&serial, "serial", "", "device serial, when more than one is connected")
	return c
}

func newPryonRemoveCmd() *cobra.Command {
	var serial string

	c := &cobra.Command{
		Use:   "remove",
		Short: "Remove the Pryon companion and the Android media bridge, back to direct ALSA",
		Long: "Removes only what `echolocal pryon install` put there. Amazon's SpeechInteractionManager,\n" +
			"native libraries and models are left untouched. The package itself goes on the next boot.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, out := cmd.Context(), cmd.OutOrStdout()
			d, err := connect(ctx, out, serial)
			if err != nil {
				return err
			}
			return render(
				ctx,
				out,
				"Removing Pryon",
				"✓ Pryon removed; echod is back on direct ALSA",
				func(report installer.Reporter) error {
					return installer.PryonRemove(ctx, d, report)
				},
			)
		},
	}
	c.Flags().StringVar(&serial, "serial", "", "device serial, when more than one is connected")
	return c
}
