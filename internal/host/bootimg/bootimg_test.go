package bootimg

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func made(size int64, cmdline string) ([]byte, Image) {
	data := make([]byte, size)
	copy(data, magic)
	copy(data[cmdlineOffset:], cmdline)

	sum := sha256.Sum256(data)
	return data, Image{
		SHA256:  hex.EncodeToString(sum[:]),
		Size:    size,
		Cmdline: "androidboot.selinux=permissive",
		Build:   13121734532,
		Device:  "biscuit_puffin",
	}
}

func TestVerifyAcceptsWhatItDescribes(t *testing.T) {
	data, img := made(4096, "bootopt=64S3 androidboot.selinux=permissive")

	if err := img.Verify("test", data); err != nil {
		t.Fatalf("Verify on its own image: %v", err)
	}
}

func TestVerifyRefusesEachWay(t *testing.T) {
	data, img := made(4096, "bootopt=64S3 androidboot.selinux=permissive")

	for _, tc := range []struct {
		name  string
		data  []byte
		about string
	}{
		{"short", data[:2048], "bytes"},
		{"long", append(append([]byte{}, data...), 0), "bytes"},
		{"wrong magic", func() []byte {
			d := append([]byte{}, data...)
			copy(d, "XNDROID!")
			return d
		}(), "does not start with"},
		{"one byte flipped", func() []byte {
			d := append([]byte{}, data...)
			d[3000] ^= 0xFF
			return d
		}(), "hashes to"},
	} {
		err := img.Verify(tc.name, tc.data)
		if err == nil {
			t.Errorf("%s: accepted", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), tc.about) {
			t.Errorf("%s: %v, want it to mention %q", tc.name, err, tc.about)
		}
	}
}

func TestVerifyRefusesAnEnforcingCmdline(t *testing.T) {
	data, img := made(4096, "bootopt=64S3,32N2,64N2")

	err := img.Verify("enforcing", data)
	if err == nil {
		t.Fatal("accepted an image with no permissive cmdline")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("%v, want it to say what is missing", err)
	}
}

func TestSupportsRefusesTheDeviceAndOnlyRemarksOnTheBuild(t *testing.T) {
	_, img := made(4096, "androidboot.selinux=permissive")

	for _, tc := range []struct {
		device, build       string
		ok, sayingSomething bool
	}{
		{"biscuit_puffin", "13121734532", true, false},
		{"biscuit_puffin", "13121734533", true, true},
		{"biscuit_puffin", "13121734531", true, false},
		{"biscuit_puffin", "", true, true},
		{"biscuit_puffin", "272.6.8.0_user_680767620", true, true},
		{"biscuit", "13121734532", false, false},
		{"tank", "13121734532", false, false},
	} {
		said, err := img.Supports(tc.device, tc.build)
		if (err == nil) != tc.ok {
			t.Errorf("Supports(%q, %q) = %v, want ok=%t", tc.device, tc.build, err, tc.ok)
		}
		if err == nil && (said != "") != tc.sayingSomething {
			t.Errorf(
				"Supports(%q, %q) said %q, want anything=%t",
				tc.device,
				tc.build,
				said,
				tc.sayingSomething,
			)
		}
	}
}

func TestPartitionIsTheSlot(t *testing.T) {
	if got := Partition("_b"); got != "boot_b" {
		t.Errorf("Partition(_b) = %q", got)
	}
	if got := Node("_a"); got != ByName+"boot_a" {
		t.Errorf("Node(_a) = %q", got)
	}
}

func TestCmdline(t *testing.T) {
	data, _ := made(4096, "bootopt=64S3 androidboot.selinux=permissive")
	if got := Cmdline(data); got != "bootopt=64S3 androidboot.selinux=permissive" {
		t.Errorf("Cmdline = %q", got)
	}
	if got := Cmdline(data[:16]); got != "" {
		t.Errorf("Cmdline of a truncated header = %q, want empty", got)
	}
}

func TestEveryImageFitsHowItIsWritten(t *testing.T) {
	for _, img := range Images {
		if img.Size%512 != 0 {
			t.Errorf("%s image size %d is not a multiple of 512", img.Device, img.Size)
		}
		for _, size := range PartitionSizes {
			if img.Size >= size {
				t.Errorf("%s image is %d bytes and the partition is %d", img.Device, img.Size, size)
			}
		}
	}
}

func TestEveryImageIsInTheTreeAndIsWhatWeSayItIs(t *testing.T) {
	for _, img := range Images {
		if img.File == "" {
			t.Errorf("%s names no file to publish", img.Device)
			continue
		}

		path := filepath.Join("../assets", img.File)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("reading %s: %v", img.Device, err)
			continue
		}
		if err := img.Verify(path, data); err != nil {
			t.Error(err)
		}

		if base := filepath.Base(img.File); base != img.Asset() {
			t.Errorf("%s is committed as %q and published at %q", img.Device, base, img.Asset())
		}
		if !strings.Contains(img.URL, "/"+Tag+"/") {
			t.Errorf(
				"%s is published at %q, which is not under the %s tag",
				img.Device,
				img.URL,
				Tag,
			)
		}
	}
}
