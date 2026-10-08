package device

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
)

const Binary = "adb"

const rcMarker = "__echolocal_rc="

const PlatformTools = "https://developer.android.com/tools/releases/platform-tools"

func Require() error {
	if _, err := exec.LookPath(Binary); err != nil {
		return fmt.Errorf("adb is not on PATH: install Android platform-tools and try again\n"+
			"  %s\n"+
			"  macOS: brew install --cask android-platform-tools\n"+
			"  Debian/Ubuntu: apt install android-sdk-platform-tools", PlatformTools)
	}
	return nil
}

type Device struct {
	serial string
}

type Error struct {
	Cmd    string
	Output string
	Code   int
}

func (e *Error) Error() string {
	out := strings.TrimSpace(e.Output)
	if out == "" {
		return fmt.Sprintf("device: %q exited %d", e.Cmd, e.Code)
	}
	return fmt.Sprintf("device: %q exited %d: %s", e.Cmd, e.Code, out)
}

type Info struct {
	Serial  string
	Model   string
	Product string
}

func (i Info) String() string {
	switch {
	case i.Model != "" && i.Product != "":
		return fmt.Sprintf("%s (%s)", i.Model, i.Product)
	case i.Model != "":
		return i.Model
	}
	return i.Serial
}

const (
	StateOnline   = "device"
	StateRecovery = "recovery"
)

var ErrUnreachable = errors.New("no device adb can drive")

func List() ([]Info, error) { return list(StateOnline) }

func ListAny() ([]Info, error) { return list(StateOnline, StateRecovery) }

func list(states ...string) ([]Info, error) {
	serials, err := listing(states...)
	if err != nil {
		return nil, err
	}

	out := make([]Info, 0, len(serials))
	for _, serial := range serials {
		out = append(out, probe(serial))
	}
	return out, nil
}

// An adb shell into some boards never returns.
const probeTimeout = 3 * time.Second

func probe(serial string) Info {
	d := &Device{serial: serial}
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	model, _ := d.run(ctx, "shell", "getprop ro.product.model")
	product, _ := d.run(ctx, "shell", "getprop ro.product.device")
	return Info{
		Serial:  serial,
		Model:   strings.TrimSpace(model),
		Product: strings.TrimSpace(product),
	}
}

func (d *Device) InRecovery() (bool, error) {
	serials, err := listing(StateRecovery)
	if err != nil {
		return false, err
	}
	return slices.Contains(serials, d.serial), nil
}

// adb devices prints "serial<tab>state" per line after a header.
func listing(states ...string) ([]string, error) {
	out, err := run(context.Background(), "devices")
	if err != nil {
		return nil, err
	}

	var serials []string
	for line := range strings.SplitSeq(out, "\n") {
		serial, state, ok := strings.Cut(strings.TrimSpace(line), "\t")
		if !ok {
			continue
		}
		for _, want := range states {
			if strings.TrimSpace(state) == want {
				serials = append(serials, serial)
			}
		}
	}
	return serials, nil
}

func Connect(serial string) (*Device, error) {
	d, err := Attach(serial)
	if err != nil {
		return nil, err
	}
	return rooted(d)
}

func Attach(serial string) (*Device, error) {
	return attach(serial, StateOnline)
}

func AttachAny(serial string) (*Device, error) {
	return attach(serial, StateOnline, StateRecovery)
}

func attach(serial string, states ...string) (*Device, error) {
	if err := Require(); err != nil {
		return nil, err
	}
	serials, err := listing(states...)
	if err != nil {
		return nil, err
	}

	switch {
	case len(serials) == 0:
		return nil, ErrUnreachable
	case serial != "":
		for _, s := range serials {
			if s == serial {
				return &Device{serial: s}, nil
			}
		}
		return nil, fmt.Errorf("device: no device with serial %q", serial)
	case len(serials) > 1:
		return nil, fmt.Errorf("device: %d devices connected, pick one with --serial: %s",
			len(serials), strings.Join(serials, ", "))
	}
	return &Device{serial: serials[0]}, nil
}

func rooted(d *Device) (*Device, error) {
	root, err := d.IsRoot()
	if err != nil {
		return nil, err
	}
	if !root {
		return nil, errors.New(
			"device: adbd is not running as root; write the boot image first (echoctl install --flash-only)",
		)
	}
	return d, nil
}

func (d *Device) Serial() string { return d.serial }

// "recovery" is TWRP, the only place the boot partition can be written.
func (d *Device) Reboot(target string) error {
	args := []string{"reboot"}
	if target != "" {
		args = append(args, target)
	}
	_, err := d.run(context.Background(), args...)
	return err
}

// The root source is /dev/root with no node: mount(8) fails, mount(2) with no source works.
func (d *Device) Remount() error {
	_, err := d.run(context.Background(), "remount")
	return err
}

// adb reports a device present a moment before it will run anything.
func (d *Device) Wait(ctx context.Context, states ...string) error {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		serials, err := listing(states...)
		if err == nil {
			for _, s := range serials {
				if s != d.serial {
					continue
				}
				if _, err := d.rawShell("true"); err == nil {
					return nil
				}
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("device: waiting for %s in %v: %w", d.serial, states, ctx.Err())
		case <-tick.C:
		}
	}
}

// sys.boot_completed is set by system_server, which echod keeps off.
func (d *Device) WaitRooted(ctx context.Context) error {
	if err := d.Wait(ctx, StateOnline); err != nil {
		return err
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		if root, err := d.IsRoot(); err == nil && root {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("device: %s did not come back as root: %w", d.serial, ctx.Err())
		case <-tick.C:
		}
	}
}

func (d *Device) WaitBooted(ctx context.Context) error {
	if err := d.Wait(ctx, StateOnline); err != nil {
		return err
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		if done, _ := d.Getprop("sys.boot_completed"); done == "1" {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("device: %s did not finish booting: %w", d.serial, ctx.Err())
		case <-tick.C:
		}
	}
}

func (d *Device) Shell(cmd string) (string, error) {
	out, code, err := d.shell(cmd)
	if err != nil {
		return out, err
	}
	if code != 0 {
		return out, &Error{Cmd: cmd, Output: out, Code: code}
	}
	return out, nil
}

func (d *Device) ShellCode(cmd string) (string, int, error) {
	return d.shell(cmd)
}

// Older adb does not propagate the shell exit status.
func (d *Device) shell(cmd string) (string, int, error) {
	raw, err := d.withRetry(func() (string, error) {
		raw, err := d.rawShell(cmd)
		if err == nil && !strings.Contains(raw, rcMarker) {
			return raw, errCutOff
		}
		return raw, err
	})
	if err != nil {
		return "", 0, fmt.Errorf("device: running %q: %w", cmd, err)
	}
	out, code, err := splitRC(raw)
	if err != nil {
		return out, 0, fmt.Errorf("device: running %q: %w", cmd, err)
	}
	return out, code, nil
}

func splitRC(raw string) (string, int, error) {
	out := strings.ReplaceAll(raw, "\r\n", "\n")

	i := strings.LastIndex(out, rcMarker)
	if i < 0 {
		return out, 0, fmt.Errorf("no %s marker in output %q", rcMarker, out)
	}
	status, trailing, _ := strings.Cut(out[i+len(rcMarker):], "\n")
	code, err := strconv.Atoi(strings.TrimSpace(status))
	if err != nil {
		return out, 0, fmt.Errorf("unparseable exit status in %q", out[i:])
	}
	text := strings.TrimSuffix(out[:i], "\n")
	if trailing = strings.TrimSuffix(trailing, "\n"); trailing != "" {
		if text != "" {
			text += "\n"
		}
		text += trailing
	}
	return text, code, nil
}

func (d *Device) Exists(path string) (bool, error) {
	_, code, err := d.ShellCode("ls -d " + quote(path))
	if err != nil {
		return false, err
	}
	return code == 0, nil
}

func (d *Device) IsSymlink(path string) (bool, error) {
	_, code, err := d.ShellCode("[ -L " + quote(path) + " ]")
	if err != nil {
		return false, err
	}
	return code == 0, nil
}

// exec-out is the only adb shell that does not translate line endings.
func (d *Device) ReadFile(path string) ([]byte, error) {
	out, err := d.withRetry(func() (string, error) {
		b, err := d.command(context.Background(), "exec-out", "cat "+quote(path)).Output()
		return string(b), err
	})
	if err != nil {
		return nil, fmt.Errorf("device: reading %s: %w", path, err)
	}
	return []byte(out), nil
}

func (d *Device) PullFile(remote, local string) error {
	if _, err := d.run(context.Background(), "pull", remote, local); err != nil {
		return fmt.Errorf("device: pulling %s: %w", remote, err)
	}
	return nil
}

func (d *Device) WriteFile(remote string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp("", "echoctl")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return d.PushFile(tmp.Name(), remote, mode)
}

// adb push widens a new file's owner bits to group and other, and leaves an overwritten one 0644.
func (d *Device) PushFile(local, remote string, mode os.FileMode) error {
	if _, err := d.withRetry(func() (string, error) {
		return d.run(context.Background(), "push", local, remote)
	}); err != nil {
		return fmt.Errorf("device: pushing %s: %w", remote, err)
	}

	_, err := d.Shell(fmt.Sprintf("chmod %o %s", mode.Perm(), quote(remote)))
	return err
}

func (d *Device) Getprop(name string) (string, error) {
	out, err := d.Shell("getprop " + quote(name))
	return strings.TrimSpace(out), err
}

func (d *Device) Setprop(name, value string) error {
	_, err := d.Shell(fmt.Sprintf("setprop %s %s", quote(name), quote(value)))
	return err
}

func (d *Device) Label(path string) (string, error) {
	out, err := d.Shell("ls -Zd " + quote(path))
	if err != nil {
		return "", err
	}
	for f := range strings.FieldsSeq(out) {
		if strings.HasPrefix(f, "u:object_r:") {
			return f, nil
		}
	}
	return "", fmt.Errorf("device: no SELinux label in %q", strings.TrimSpace(out))
}

func (d *Device) Chcon(label, path string) error {
	_, err := d.Shell(fmt.Sprintf("chcon %s %s", quote(label), quote(path)))
	return err
}

// id -u is unsupported on the device.
func (d *Device) IsRoot() (bool, error) {
	out, err := d.Shell("id")
	if err != nil {
		return false, err
	}
	return strings.Contains(out, "uid=0("), nil
}

func (d *Device) Uptime() (float64, error) {
	out, err := d.Shell("cat /proc/uptime")
	if err != nil {
		return 0, err
	}
	first, _, _ := strings.Cut(strings.TrimSpace(out), " ")
	v, err := strconv.ParseFloat(first, 64)
	if err != nil {
		return 0, fmt.Errorf("device: unparseable uptime %q", out)
	}
	return v, nil
}

func (d *Device) command(ctx context.Context, args ...string) *exec.Cmd {
	if d.serial != "" {
		args = append([]string{"-s", d.serial}, args...)
	}
	return exec.CommandContext(ctx, Binary, args...)
}

func (d *Device) run(ctx context.Context, args ...string) (string, error) {
	return output(d.command(ctx, args...))
}

var errCutOff = errors.New("the shell ended before its exit status")

func Dropped(err error) bool {
	if errors.Is(err, errCutOff) {
		return true
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		return false
	}
	if exit.ExitCode() == 255 {
		return true
	}
	msg := err.Error()
	return exit.ExitCode() == 1 &&
		(strings.Contains(msg, "adb: device '") && strings.Contains(msg, "' not found") ||
			strings.Contains(msg, "device offline") ||
			strings.Contains(msg, "no devices/emulators found"))
}

func (d *Device) withRetry(do func() (string, error)) (string, error) {
	out, err := do()
	if !Dropped(err) {
		return out, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), backTimeout)
	defer cancel()

	if err := d.back(ctx); err != nil {
		return out, err
	}
	return do()
}

func (d *Device) rawShell(cmd string) (string, error) {
	return d.run(context.Background(), "shell", cmd+"; echo "+rcMarker+"$?")
}

func (d *Device) back(ctx context.Context) error {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		if raw, err := d.rawShell("true"); err == nil && strings.Contains(raw, rcMarker) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("device: %s went away and did not come back: %w", d.serial, ctx.Err())
		case <-tick.C:
		}
	}
}

const backTimeout = 5 * time.Second

func run(ctx context.Context, args ...string) (string, error) {
	return output(exec.CommandContext(ctx, Binary, args...))
}

// adb explains refusals on stderr.
func output(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	text := string(out)
	if err != nil {
		said := strings.Join(strings.Fields(text), " ")
		return text, fmt.Errorf("%s: %w: %s", strings.Join(cmd.Args, " "), err, said)
	}
	return text, nil
}

func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
