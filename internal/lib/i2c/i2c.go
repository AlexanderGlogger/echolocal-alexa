package i2c

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// I2C_SLAVE fails with EBUSY on a driver-owned device; I2C_SLAVE_FORCE claims it anyway.
	ioctlSlave      = 0x0703
	ioctlSlaveForce = 0x0706
)

type Bus struct {
	f    *os.File
	addr uint8
}

// Writes to a driver-owned device race the driver.
func Open(n int, addr uint8, force bool) (*Bus, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", n), os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	req := uintptr(ioctlSlave)
	if force {
		req = ioctlSlaveForce
	}
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), req, uintptr(addr)); e != 0 {
		_ = f.Close()
		return nil, fmt.Errorf("set slave 0x%02x on i2c-%d: %w", addr, n, e)
	}
	return &Bus{f: f, addr: addr}, nil
}

func (b *Bus) Close() error { return b.f.Close() }

func (b *Bus) ReadReg(reg uint8) (uint8, error) {
	if _, err := b.f.Write([]byte{reg}); err != nil {
		return 0, err
	}
	var buf [1]byte
	if _, err := b.f.Read(buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (b *Bus) WriteReg(reg, val uint8) error {
	_, err := b.f.Write([]byte{reg, val})
	return err
}

func (b *Bus) Dump(first, last uint8) ([]uint8, error) {
	if last < first {
		return nil, fmt.Errorf("i2c: bad range %d..%d", first, last)
	}
	out := make([]uint8, 0, int(last-first)+1)
	for r := int(first); r <= int(last); r++ {
		v, err := b.ReadReg(uint8(r))
		if err != nil {
			v = 0xFF
		}
		out = append(out, v)
	}
	return out, nil
}

var _ = unsafe.Pointer(nil)
