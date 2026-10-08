package installer

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ygelfand/libcountertop/pkg/host/adb"

	"github.com/ygelfand/echolocal/internal/host/bootimg"
)

// Amazon's adbd drops privileges however ro.secure is set.

const (
	recoveryTimeout = 90 * time.Second
	androidTimeout  = 3 * time.Minute
)

// /tmp in recovery is a ramdisk.
const remoteImage = "/tmp/echolocal-boot.img"

var flashSteps = []step{
	{"check device", checkState},
	{"check boot image", checkImage},
	{"check approval", checkApproval},
	{"reboot to recovery", bootRecovery},
	{"check target partition", checkPartition},
	{"write the boot image", writeImage},
	{"verify what was written", verifyImage},
	{"disable dm-verity", disableVerity},
	{"patch the root filesystem", patchSystem},
	{"clear the saved usb config", clearUSBConfig},
	{"reboot to android", bootAndroid},
	{"confirm root and permissive", confirmRoot},
}

type state struct {
	device     string
	build      string
	rooted     bool
	recovery   bool
	permissive bool
	enforcing  string

	slot string
}

func (s state) ready() bool { return s.rooted && s.permissive && !s.recovery }

func (s state) String() string {
	return fmt.Sprintf("root=%t selinux=%s", s.rooted, strings.TrimSpace(s.enforcing))
}

type BootState struct {
	Ready bool

	Summary string

	Partition string
}

func Probe(d *adb.Device) (BootState, error) {
	s, err := probe(d)
	return BootState{
		Ready:     s.ready(),
		Summary:   s.String(),
		Partition: bootimg.Partition(s.slot),
	}, err
}

func probe(d *adb.Device) (state, error) {
	var s state
	var err error

	if s.recovery, err = d.InRecovery(); err != nil {
		return s, err
	}
	if s.device, err = d.Getprop("ro.product.device"); err != nil {
		return s, err
	}
	if s.slot, err = d.Getprop("ro.boot.slot_suffix"); err != nil {
		return s, err
	}
	if s.build, err = d.Getprop("ro.build.version.incremental"); err != nil {
		return s, err
	}
	if s.rooted, err = d.IsRoot(); err != nil {
		return s, err
	}

	s.enforcing, _ = d.Shell("getenforce")
	s.permissive = !strings.EqualFold(strings.TrimSpace(s.enforcing), "enforcing")
	return s, nil
}

func checkState(r *run) (string, bool, error) {
	s, err := probe(r.d)
	if err != nil {
		return "", false, err
	}
	r.state = s

	return fmt.Sprintf("%s, build %s, %s", s.device, s.build, s), false, nil
}

func checkImage(r *run) (string, bool, error) {
	if detail, skip := r.unflashed(); skip {
		return detail, true, nil
	}
	if detail, skip := r.done(); skip {
		return detail, true, nil
	}
	if len(r.cfg.BootImage) == 0 {
		return "", false, errors.New("no boot image given")
	}
	if err := r.cfg.Profile.Boot.Verify(r.cfg.BootImageFrom, r.cfg.BootImage); err != nil {
		return "", false, err
	}
	detail := fmt.Sprintf(
		"%s, %d bytes, %s",
		r.cfg.BootImageFrom,
		r.cfg.Profile.Boot.Size,
		r.cfg.Profile.Boot.SHA256[:12],
	)

	// getprop in recovery describes the recovery image.
	if r.state.recovery {
		return detail, false, nil
	}

	untested, err := r.cfg.Profile.Boot.Supports(r.state.device, r.state.build)
	if err != nil {
		return "", false, err
	}
	if untested != "" {
		detail += "; " + untested
	}
	return detail, false, nil
}

func checkApproval(r *run) (string, bool, error) {
	if detail, skip := r.done(); skip {
		return detail, true, nil
	}
	if !r.cfg.Approved {
		return "", false, errors.New("overwriting the boot partition was not approved")
	}
	return "approved", false, nil
}

func (r *run) done() (string, bool) {
	if r.state.ready() && !r.cfg.Reflash {
		return "already root and permissive", true
	}
	return "", false
}

func (r *run) unflashed() (string, bool) {
	if r.cfg.Profile.Boot.SHA256 == "" {
		return "this board keeps its own boot image", true
	}
	return "", false
}

func (r *run) settle() error {
	ctx, cancel := context.WithTimeout(r.ctx, recoveryTimeout)
	defer cancel()

	return r.d.Wait(ctx, adb.StateRecovery)
}

const stageTries = 3

func (r *run) stage() error {
	tries := make([]string, 0, stageTries)

	for try := range stageTries {
		if err := r.settle(); err != nil {
			return err
		}
		if err := r.d.WriteFile(
			remoteImage,
			r.cfg.BootImage,
			0o644,
		); err != nil &&
			try == stageTries-1 {
			return err
		}
		if err := r.settle(); err != nil {
			return err
		}

		staged, err := sha256Of(r.d, "cat "+remoteImage)
		if err != nil && try == stageTries-1 {
			return err
		}
		if staged == r.cfg.Profile.Boot.SHA256 {
			return nil
		}
		tries = append(tries, fmt.Sprintf("%s bytes %s", sizeOf(r.d, remoteImage), staged))
	}

	return fmt.Errorf("the staged image never matched. tries: %s. want: %d bytes %s",
		strings.Join(tries, "; "), r.cfg.Profile.Boot.Size, r.cfg.Profile.Boot.SHA256)
}

func sizeOf(d *adb.Device, remote string) string {
	out, err := d.Shell("wc -c < " + remote)
	if err != nil {
		return "?"
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return "?"
	}
	return fields[0]
}

// A device that boots with persist.sys.usb.config=mtp brings up no adb.
func clearUSBConfig(r *run) (string, bool, error) {
	const path = "/data/property/persist.sys.usb.config"

	there, err := r.d.Exists(path)
	if err != nil {
		return "", false, err
	}
	if !there {
		return "none saved", true, nil
	}
	if _, err := r.d.Shell("rm -f " + path); err != nil {
		return "", false, err
	}
	return "removed", false, nil
}

func bootRecovery(r *run) (string, bool, error) {
	if detail, skip := r.done(); skip {
		return detail, true, nil
	}

	if !r.state.recovery {
		if err := r.d.Reboot(adb.StateRecovery); err != nil {
			return "", false, err
		}
		ctx, cancel := context.WithTimeout(r.ctx, recoveryTimeout)
		defer cancel()

		if err := r.d.Wait(ctx, adb.StateRecovery); err != nil {
			return "", false, err
		}
	}

	root, err := r.d.IsRoot()
	if err != nil {
		return "", false, err
	}
	if !root {
		if err := r.d.Reboot(""); err != nil {
			return "", false, fmt.Errorf(
				"recovery is not running adbd as root, and rebooting back failed: %w",
				err,
			)
		}
		return "", false, errors.New(
			"recovery is not running adbd as root: this needs TWRP installed as the recovery partition; rebooting back to Android",
		)
	}
	return "root recovery", false, nil
}

func checkPartition(r *run) (string, bool, error) {
	if detail, skip := r.unflashed(); skip {
		return detail, true, nil
	}
	if detail, skip := r.done(); skip {
		return detail, true, nil
	}

	target, err := r.partition()
	if err != nil {
		return "", false, err
	}

	node, err := r.d.Shell("readlink -f " + bootimg.Node(r.state.slot))
	if err != nil {
		return "", false, err
	}
	node = strings.TrimSpace(node)

	raw, err := r.d.Shell("blockdev --getsize64 " + bootimg.Node(r.state.slot))
	if err != nil {
		return "", false, err
	}
	size, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return "", false, fmt.Errorf("%s resolves to %s, whose size reads as %q: refusing to write",
			target, node, strings.TrimSpace(raw))
	}
	if !bootimg.KnownPartition(size) {
		return "", false, fmt.Errorf(
			"%s resolves to %s of %d bytes, want one of %v: refusing to write",
			target,
			node,
			size,
			bootimg.PartitionSizes,
		)
	}
	return fmt.Sprintf("%s → %s, %d bytes", target, node, size), false, nil
}

func (r *run) partition() (string, error) {
	if r.board().Slotted && r.state.slot == "" {
		return "", errUnknownSlot
	}
	return bootimg.Partition(r.state.slot), nil
}

var errUnknownSlot = errors.New("device names no boot slot in ro.boot.slot_suffix")

func writeImage(r *run) (string, bool, error) {
	if detail, skip := r.unflashed(); skip {
		return detail, true, nil
	}
	if detail, skip := r.done(); skip {
		return detail, true, nil
	}

	// Recovery restarts its adbd shortly after it comes up, cutting off whatever is running.
	if err := r.stage(); err != nil {
		return "", false, err
	}

	// busybox builds differ on whether dd accepts conv=fsync.
	if _, err := r.d.Shell(
		fmt.Sprintf("dd if=%s of=%s bs=1048576 && sync", remoteImage, bootimg.Node(r.state.slot)),
	); err != nil {
		return "", false, err
	}
	return "written", false, nil
}

func verifyImage(r *run) (string, bool, error) {
	if detail, skip := r.unflashed(); skip {
		return detail, true, nil
	}
	if detail, skip := r.done(); skip {
		return detail, true, nil
	}

	if err := r.settle(); err != nil {
		return "", false, err
	}

	const block = 512
	if r.cfg.Profile.Boot.Size%block != 0 {
		return "", false, fmt.Errorf(
			"image size %d is not a multiple of %d",
			r.cfg.Profile.Boot.Size,
			block,
		)
	}
	got, err := sha256Of(r.d, fmt.Sprintf("dd if=%s bs=%d count=%d 2>/dev/null",
		bootimg.Node(r.state.slot), block, r.cfg.Profile.Boot.Size/block))
	if err != nil {
		return "", false, err
	}
	if got != r.cfg.Profile.Boot.SHA256 {
		return "", false, fmt.Errorf(
			"the partition hashes to %s, want %s: left in recovery, re-run to write it again",
			got,
			r.cfg.Profile.Boot.SHA256,
		)
	}
	return got[:12] + " matches", false, nil
}

func bootAndroid(r *run) (string, bool, error) {
	if detail, skip := r.done(); skip {
		return detail, true, nil
	}
	if err := r.d.Reboot(""); err != nil {
		return "", false, err
	}

	ctx, cancel := context.WithTimeout(r.ctx, androidTimeout)
	defer cancel()

	if !r.board().SignalsBoot {
		if err := r.d.WaitRooted(ctx); err != nil {
			return "", false, err
		}
		return "root adb is back", false, nil
	}

	if err := r.d.WaitBooted(ctx); err != nil {
		return "", false, err
	}
	return "booted", false, nil
}

func confirmRoot(r *run) (string, bool, error) {
	s, err := probe(r.d)
	if err != nil {
		return "", false, err
	}
	if !s.ready() {
		return "", false, fmt.Errorf("still %s", s)
	}
	return s.String(), false, nil
}

// Recovery has sha256sum.
func sha256Of(d *adb.Device, cmd string) (string, error) {
	out, err := d.Shell(cmd + " | sha256sum")
	if err != nil {
		return "", err
	}
	sum, _, _ := strings.Cut(strings.TrimSpace(out), " ")
	if len(sum) != 64 {
		return "", fmt.Errorf("unreadable sha256 in %q", strings.TrimSpace(out))
	}
	return sum, nil
}
