package installer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ygelfand/libcountertop/pkg/host/adb"

	"github.com/ygelfand/echolocal/internal/layout"
)

// Amazon's native Pryon wake detector, on an Echo Dot 2 still running Fire OS 5.
//
// The main install targets Fire OS 6, which on biscuit has no Android framework: no package manager
// to scan a privileged APK, no AudioRecord to share Pryon's capture through. Pryon needs both, so it is
// installed separately, onto a Dot that already runs echod on Fire OS 5, and touches nothing the main
// install owns beyond stopping and starting echod.
//
// Amazon's libraries, APK and models stay on the Dot and are discovered there. What is installed is
// EchoLocal's own: the wake-only companion APK and the app_process media bridge.

// PryonSDK is the only firmware Pryon is installed onto.
const PryonSDK = "22"

// PryonConfig is what a Pryon install needs from the caller.
type PryonConfig struct {
	// APK is the companion, and Helper the media bridge, both built from android/.
	APK    []byte
	Helper []byte

	// Echod, when set, replaces the installed binary in the same run. It has to be a build carrying
	// the Android media bridge, since the one already there may predate it.
	Echod []byte
}

type pryonPaths struct {
	amazonAPK string
	alexa     string
	aed       string
}

// pryonRun is a run with what only a Pryon install has. Its steps are methods so that they fit the
// step table while reaching both halves.
type pryonRun struct {
	*run
	pcfg  PryonConfig
	paths pryonPaths
}

// A newly installed SpeechInteractionManager can spend over 30 seconds in dex2oat before Pryon can
// load its native model. Keep this bounded, but allow enough time for the first boot on the Dot.
const pryonReadyTimeout = 90 * time.Second

// PryonInstall stages the companion and the media bridge, and echod when one is given. It reports
// whether Android still has to scan the companion, which only happens at boot: FinalizePryon cannot
// run until it has.
func PryonInstall(
	ctx context.Context,
	d *adb.Device,
	cfg PryonConfig,
	report Reporter,
) (bool, error) {
	p := &pryonRun{run: &run{d: d, ctx: ctx, cfg: Config{Echod: cfg.Echod}}, pcfg: cfg}
	if err := execute(ctx, pryonInstallSteps(p), p.run, report); err != nil {
		return false, err
	}
	scanned, err := PryonScanned(d)
	if err != nil {
		return false, err
	}
	return p.reboot || !scanned, nil
}

// echod stops before its binary is written. While it runs, the old binary stays allocated after adb
// unlinks it, and a Fire OS 5 /system has no room for two current builds.
func pryonInstallSteps(p *pryonRun) []step {
	return []step{
		{"check device", p.checkDevice},
		{"inspect Pryon firmware", p.inspect},
		{"install Android media bridge", p.installHelper},
		{"remount /system rw", remountRW},
		{"install Pryon companion", p.installAPK},
		{"stop echod", stopService},
		{"install echod", p.installEchod},
		{"remount /system ro", p.remountRO},
		{"start echod", startService},
	}
}

// FinalizePryon runs once Android has scanned the companion. It pins the package UID the bridge
// authenticates, hands the companion the firmware paths and verifies both the ESPHome-facing runtime
// and the native detector.
func FinalizePryon(ctx context.Context, d *adb.Device, report Reporter) error {
	p := &pryonRun{run: &run{d: d, ctx: ctx}}
	steps := []step{
		{"rediscover Pryon firmware", p.inspect},
		{"record Pryon package UID", recordPryonUID},
		{"configure Pryon detector", p.configure},
		{"stop echod", stopService},
		{"start echod with Alexa capability", startService},
		{"verify EchoLocal ESPHome service", verifyESPHome},
		{"verify Pryon detector", verifyPryon},
	}
	return execute(ctx, steps, p.run, report)
}

// PryonRemove takes the companion and the bridge off and leaves echod on direct ALSA. Package removal
// settles on the next boot.
func PryonRemove(ctx context.Context, d *adb.Device, report Reporter) error {
	p := &pryonRun{run: &run{d: d, ctx: ctx}}
	steps := []step{
		{"stop echod", stopService},
		{"stop Pryon companion", stopPryon},
		{"remount /system rw", remountRW},
		{"remove Pryon companion", removePryonAPK},
		{"remove Android media bridge", removeAndroidMedia},
		{"remount /system ro", p.remountRO},
		{"start echod", startService},
	}
	return execute(ctx, steps, p.run, report)
}

// PryonState is what a device says about its Pryon install.
type PryonState struct {
	AndroidMedia bool
	Installed    bool
	Configured   bool
	APIListening bool
}

// ReadPryonState reads it.
func ReadPryonState(d *adb.Device) PryonState {
	var s PryonState
	s.AndroidMedia, _ = d.Exists(layout.AndroidMediaJar)
	s.Installed, _ = d.Exists(layout.PryonAPK)
	s.Configured, _ = d.Exists(layout.PryonUIDPath)
	s.APIListening, _ = apiListening(d)
	return s
}

func (p *pryonRun) checkDevice(r *run) (string, bool, error) {
	if len(p.pcfg.APK) == 0 {
		return "", false, errors.New("no Pryon companion APK given")
	}
	if len(p.pcfg.Helper) == 0 {
		return "", false, errors.New("no Android media helper given")
	}

	sdk, err := r.d.Getprop("ro.build.version.sdk")
	if err != nil {
		return "", false, err
	}
	model, _ := r.d.Getprop("ro.product.device")
	if model != "biscuit" || sdk != PryonSDK {
		return "", false, fmt.Errorf(
			"device is %s on SDK %s; Pryon needs an Echo Dot 2 (biscuit) on Fire OS 5 (SDK %s)",
			model,
			sdk,
			PryonSDK,
		)
	}

	have, err := r.d.Exists(layout.Binary)
	if err != nil {
		return "", false, err
	}
	if !have && len(p.pcfg.Echod) == 0 {
		return "", false, fmt.Errorf(
			"echod is not installed at %s; install EchoLocal first",
			layout.Binary,
		)
	}
	return fmt.Sprintf("%s, SDK %s, %s", model, sdk, r.d.Serial()), false, nil
}

func (p *pryonRun) inspect(r *run) (string, bool, error) {
	for _, library := range []string{
		"/system/lib/libpryon.so",
		"/system/lib/libwakewordserver_jni.so",
		"/system/lib/libaudiostream.so",
		"/system/lib/libaudiostream_jni.so",
	} {
		have, err := r.d.Exists(library)
		if err != nil {
			return "", false, err
		}
		if !have {
			return "", false, fmt.Errorf("required firmware library is missing: %s", library)
		}
	}

	path, err := packageAPK(r.d, "amazon.speech.sim")
	if err != nil {
		return "", false, fmt.Errorf("locating SpeechInteractionManager: %w", err)
	}

	language, _ := r.d.Getprop("persist.sys.language")
	country, _ := r.d.Getprop("persist.sys.country")
	locales := []string{}
	if language != "" && country != "" {
		locales = append(locales, language+"-"+country)
	}
	if language != "" {
		locales = append(locales, language)
	}
	// Fire OS often reports en-GB while the physical model is reached through an en-GB symlink.
	locales = append(locales, "en-GB", "en-US")

	paths := pryonPaths{amazonAPK: path}
	var searched []string
	for _, locale := range unique(locales) {
		candidate := "/system/local/models/keyword/" + locale + "/ALEXA/pryon.manifest"
		searched = append(searched, candidate)
		if have, _ := r.d.Exists(candidate); have {
			paths.alexa = candidate
			break
		}
	}
	if paths.alexa == "" {
		found, _ := r.d.Shell(
			"find /system/local/models/keyword -path '*/ALEXA/pryon.manifest' 2>/dev/null",
		)
		paths.alexa = firstLine(found)
	}
	if paths.alexa == "" {
		return "", false, fmt.Errorf("unable to locate the Alexa Pryon manifest; searched: %s",
			strings.Join(searched, ", "))
	}

	paths.aed = "/system/local/models/AED/pryon.manifest"
	if have, _ := r.d.Exists(paths.aed); !have {
		found, _ := r.d.Shell("find /system/local/models -path '*/AED/pryon.manifest' 2>/dev/null")
		paths.aed = firstLine(found)
	}
	if paths.aed == "" {
		return "", false, errors.New(
			"unable to locate the required AED Pryon manifest under /system/local/models",
		)
	}

	p.paths = paths
	return fmt.Sprintf(
		"SIM=%s, Alexa=%s, AED=%s",
		paths.amazonAPK,
		paths.alexa,
		paths.aed,
	), false, nil
}

// packageAPK resolves an installed package even after `pm hide`. Fire OS 5 makes `pm path` return
// exit 1 for a hidden package, while `pm list packages -f -u` still reports the system APK. Pryon is
// finalized after a reboot with Amazon's audio packages hidden, so both views are required.
func packageAPK(d *adb.Device, name string) (string, error) {
	direct, directCode, err := d.ShellCode("pm path " + name)
	if err != nil {
		return "", err
	}
	if directCode == 0 {
		if path := packagePath(direct, name); path != "" {
			return path, nil
		}
	}

	all, listCode, err := d.ShellCode("pm list packages -f -u " + name)
	if err != nil {
		return "", err
	}
	if listCode == 0 {
		if path := packagePath(all, name); path != "" {
			return path, nil
		}
	}
	return "", fmt.Errorf(
		"package %s reported no APK path (pm path exit %d, hidden-package list exit %d)",
		name,
		directCode,
		listCode,
	)
}

func packagePath(output, name string) string {
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "package:") {
			continue
		}
		value := strings.TrimPrefix(line, "package:")
		if path, packageName, found := strings.Cut(value, "="); found {
			if strings.TrimSpace(packageName) != name {
				continue
			}
			return strings.TrimSpace(path)
		}
		if value != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func unique(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(line)
}

func (p *pryonRun) installHelper(r *run) (string, bool, error) {
	if _, err := r.d.Shell("mkdir -p " + layout.StateDir); err != nil {
		return "", false, err
	}
	if same, err := sameRemote(r.d, layout.AndroidMediaJar, p.pcfg.Helper); err != nil {
		return "", false, err
	} else if same {
		return "already installed", true, nil
	}
	if err := r.d.WriteFile(layout.AndroidMediaJar, p.pcfg.Helper, 0o644); err != nil {
		return "", false, err
	}
	return layout.AndroidMediaJar, false, nil
}

func (p *pryonRun) installAPK(r *run) (string, bool, error) {
	if same, err := sameRemote(r.d, layout.PryonAPK, p.pcfg.APK); err != nil {
		return "", false, err
	} else if same {
		return "already installed", true, nil
	}
	// Android may still have the previous APK and its native libraries mapped. Replacing that file
	// underneath a live privileged process can keep /system busy and make the safety remount to
	// read-only fail. A fresh install is harmlessly force-stopped too.
	if _, err := r.d.Shell("am force-stop " + layout.PryonPackage); err != nil {
		return "", false, fmt.Errorf("stopping the existing Pryon companion: %w", err)
	}
	if _, err := r.d.Shell("mkdir -p " + layout.PryonDir); err != nil {
		return "", false, err
	}
	if err := r.d.WriteFile(layout.PryonAPK, p.pcfg.APK, 0o644); err != nil {
		return "", false, err
	}
	if err := r.d.Chcon(layout.OurLabel, layout.PryonAPK); err != nil {
		return "", false, err
	}
	r.reboot = true
	return layout.PryonAPK + " (effective after reboot)", false, nil
}

func (p *pryonRun) installEchod(r *run) (string, bool, error) {
	if len(p.pcfg.Echod) == 0 {
		return "keeping the installed binary", true, nil
	}
	if same, err := sameRemote(r.d, layout.Binary, p.pcfg.Echod); err != nil {
		return "", false, err
	} else if same {
		return "already installed", true, nil
	}
	return installBinary(r)
}

// remountRO is Fire OS 5's own remount. The main install goes through echod's, which has to cope with
// Fire OS 6 as well and may not be in the binary that is installed here.
func (p *pryonRun) remountRO(r *run) (string, bool, error) {
	_, err := r.d.Shell("mount -o remount,ro /system")
	return "", false, err
}

func sameRemote(d *adb.Device, path string, want []byte) (bool, error) {
	have, err := d.Exists(path)
	if err != nil || !have {
		return false, err
	}
	got, err := d.ReadFile(path)
	if err != nil {
		return false, err
	}
	return bytes.Equal(got, want), nil
}

func stopPryon(r *run) (string, bool, error) {
	have, err := r.d.Exists(layout.PryonAPK)
	if err != nil || !have {
		return "not installed", true, err
	}
	_, err = r.d.Shell("am force-stop " + layout.PryonPackage)
	return layout.PryonPackage, false, err
}

func removePryonAPK(r *run) (string, bool, error) {
	have, err := r.d.Exists(layout.PryonDir)
	if err != nil || !have {
		return "not installed", true, err
	}
	_, err = r.d.Shell("rm -rf " + layout.PryonDir)
	return layout.PryonDir + " (package removal settles after reboot)", false, err
}

func removeAndroidMedia(r *run) (string, bool, error) {
	paths := layout.AndroidMediaJar + " " + layout.PryonUIDPath
	_, err := r.d.Shell("rm -f " + paths)
	return paths, false, err
}

var userIDPattern = regexp.MustCompile(`(?m)\buserId=(\d+)\b`)

// PryonScanned reports whether Android's package manager has loaded the privileged companion.
// A newly copied system APK is not visible until the next boot.
func PryonScanned(d *adb.Device) (bool, error) {
	dump, code, err := d.ShellCode("dumpsys package " + layout.PryonPackage)
	if err != nil {
		return false, err
	}
	return code == 0 && userIDPattern.MatchString(dump), nil
}

func recordPryonUID(r *run) (string, bool, error) {
	dump, err := r.d.Shell("dumpsys package " + layout.PryonPackage)
	if err != nil {
		return "", false, err
	}
	match := userIDPattern.FindStringSubmatch(dump)
	if len(match) != 2 {
		return "", false, fmt.Errorf("package %s has no userId after reboot", layout.PryonPackage)
	}
	uid, err := strconv.Atoi(match[1])
	if err != nil || uid < 10000 {
		return "", false, fmt.Errorf(
			"package %s reported invalid UID %q",
			layout.PryonPackage,
			match[1],
		)
	}
	want := []byte(strconv.Itoa(uid) + "\n")
	if same, err := sameRemote(r.d, layout.PryonUIDPath, want); err != nil {
		return "", false, err
	} else if same {
		return strconv.Itoa(uid), true, nil
	}
	if err := r.d.WriteFile(layout.PryonUIDPath, want, 0o600); err != nil {
		return "", false, err
	}
	return strconv.Itoa(uid), false, nil
}

func (p *pryonRun) configure(r *run) (string, bool, error) {
	cmd := fmt.Sprintf(
		"am startservice -n %s/.PryonDetectorService --es amazon_apk %s --es alexa_model %s --es aed_model %s",
		layout.PryonPackage,
		p.paths.amazonAPK,
		p.paths.alexa,
		p.paths.aed,
	)
	out, err := r.d.Shell(cmd)
	if err != nil {
		return "", false, err
	}
	return strings.TrimSpace(out), false, nil
}

func apiListening(d *adb.Device) (bool, error) {
	port := fmt.Sprintf("%04X", layout.Port)
	_, code, err := d.ShellCode(
		"cat /proc/net/tcp /proc/net/tcp6 2>/dev/null | grep ':" + port + " '",
	)
	return code == 0, err
}

func verifyESPHome(r *run) (string, bool, error) {
	name, err := ReadName(r.d)
	if err != nil || name == "" {
		return "", false, fmt.Errorf("EchoLocal device name is unavailable: %w", err)
	}
	mac, err := r.d.Shell("cat " + layout.MACPath)
	if err != nil || layout.MAC(mac) == "" {
		return "", false, fmt.Errorf("EchoLocal device MAC is unavailable: %w", err)
	}
	// startService proves that init launched this binary, not that all of its components have reached
	// their listeners. The Android media helper can take several seconds to create AudioRecord and
	// AudioTrack after a cold boot, so wait for the resident state and port rather than racing them.
	readyDeadline := time.Now().Add(30 * time.Second)
	var state string
	var listening bool
	for time.Now().Before(readyDeadline) {
		state, err = r.d.Getprop(layout.StateProp)
		if err != nil {
			return "", false, err
		}
		if state == "resident" {
			if listening, err = apiListening(r.d); err != nil {
				return "", false, err
			}
			if listening {
				break
			}
		}
		select {
		case <-r.ctx.Done():
			return "", false, r.ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	if state != "resident" || !listening {
		return "", false, fmt.Errorf(
			"ESPHome native API not ready within 30s (echod state=%q, tcp/%d listening=%t)",
			state,
			layout.Port,
			listening,
		)
	}

	mdnsDeadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(mdnsDeadline) {
		logs, _ := r.d.Shell("logcat -d -s echolocal:I '*:S'")
		if mdnsAdvertised(logs) {
			return fmt.Sprintf(
				"%s, %s on tcp/%d with mDNS",
				name,
				layout.MAC(mac),
				layout.Port,
			), false, nil
		}
		select {
		case <-r.ctx.Done():
			return "", false, r.ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return "", false, errors.New(
		"ESPHome API is listening but mDNS discovery was not advertised within 15s; verify Wi-Fi before completing installation",
	)
}

// mdnsAdvertised reads either generation of echod: a run without --echod finalizes whatever binary is
// already installed. Only the line for a published record carries its addresses.
func mdnsAdvertised(logs string) bool {
	return strings.Contains(logs, "advertising what=esphome addrs=") ||
		strings.Contains(logs, "advertising over mdns")
}

func verifyPryon(r *run) (string, bool, error) {
	deadline := time.Now().Add(pryonReadyTimeout)
	var logs string
	for time.Now().Before(deadline) {
		// Fire OS routes privileged-app logs to amazon_main rather than the default buffer. `-b all`
		// keeps verification independent of which Android UID emitted each half of the handshake.
		poc, _ := r.d.Shell("logcat -b all -d -s EchoLocalPryon:I '*:S'")
		helper, _ := r.d.Shell("logcat -b all -d -s echolocal-helper:I '*:S'")
		echo, _ := r.d.Shell("logcat -b all -d -s echolocal:I '*:S'")
		logs = "Pryon: " + lastLines(poc, 6) + " | Helper: " + lastLines(helper, 4) +
			" | EchoLocal: " + lastLines(echo, 4)
		if pryonRuntimeReady(poc, helper, echo) {
			return "native detector and shared microphone ready; Alexa advertised", false, nil
		}
		select {
		case <-r.ctx.Done():
			return "", false, r.ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return "", false, fmt.Errorf(
		"the Pryon detector and its shared live microphone did not report ready within %s; recent state: %s",
		pryonReadyTimeout,
		strings.TrimSpace(logs),
	)
}

func pryonRuntimeReady(poc, helper, echo string) bool {
	echoReady := strings.Contains(echo, "pryon=1") ||
		strings.Contains(echo, "id=pryon_alexa engine=pryon")
	return strings.Contains(poc, "PRYON_READY") &&
		strings.Contains(poc, "PRYON_CAPTURE_ACTIVE") &&
		strings.Contains(poc, "PRYON_SHARED_PCM_FIRST_FRAME") &&
		strings.Contains(helper, "shared Pryon capture first frame") && echoReady
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}
