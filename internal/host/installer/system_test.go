package installer

import "testing"

func TestVendorUnder(t *testing.T) {
	for name, tc := range map[string]struct{ at, link, want string }{
		"absolute link in recovery": {"/system_root", "/system/vendor", "/system_root/system/vendor"},
		"absolute link on a device": {"", "/system/vendor", "/system/vendor"},
		"a directory, not a link":   {"/system_root", "", "/system_root/vendor"},
		"a directory on a device":   {"", "", "/vendor"},
		"relative link in recovery": {"/system_root", "system/vendor", "/system_root/system/vendor"},
	} {
		if got := vendorUnder(tc.at, tc.link); got != tc.want {
			t.Errorf("%s: vendorUnder(%q, %q) = %q, want %q", name, tc.at, tc.link, got, tc.want)
		}
	}
}

func TestSetProp(t *testing.T) {
	for name, tc := range map[string]struct{ in, want string }{
		"replaces in place": {
			"ro.a=1\nro.build.version.number=7\nro.b=2\n",
			"ro.a=1\nro.build.version.number=9\nro.b=2\n",
		},
		"appends when missing": {
			"ro.a=1\nro.b=2\n",
			"ro.a=1\nro.b=2\nro.build.version.number=9\n",
		},
		"appends to a file with no trailing newline": {
			"ro.a=1",
			"ro.a=1\nro.build.version.number=9\n",
		},
		"leaves a key it only prefixes": {
			"ro.build.version.number.extra=1\n",
			"ro.build.version.number.extra=1\nro.build.version.number=9\n",
		},
	} {
		if got := string(setProp([]byte(tc.in), "ro.build.version.number", "9")); got != tc.want {
			t.Errorf("%s:\n got %q\nwant %q", name, got, tc.want)
		}
	}
}
