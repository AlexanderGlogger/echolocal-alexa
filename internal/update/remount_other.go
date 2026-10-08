//go:build !linux

package update

import "errors"

func remount(bool) error { return errors.New("update: remounting only works on the device") }

func room(int64) error { return nil }
