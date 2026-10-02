//go:build board_checkers || board_cronos || board_rook

package ble

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// Node is MediaTek's HCI character device.
const Node = "/dev/stpbt"

// answers is how long the chip is given to say anything. One number everywhere: it answers in
// microseconds or not at all, and a longer wait has never once been the fix.
const answers = 2 * time.Second

// switched is how long the line is read after a reset, for whatever the controller announces.
const switched = 300 * time.Millisecond

// BringUp opens the radio and resets it, leaving it speaking HCI.
func BringUp() (*Port, error) {
	p, err := OpenNode(Node)
	if err != nil {
		return nil, err
	}
	if err := Reset(p); err != nil {
		p.Close()
		return nil, err
	}
	return p, nil
}

// Local is what the controller says about itself once it is running the firmware.
type Local struct {
	HCIVersion    byte
	HCIRevision   uint16
	LMPVersion    byte
	Manufacturer  uint16
	LMPSubversion uint16
}

func (l Local) String() string {
	return fmt.Sprintf("hci %d rev %d, lmp %d subversion %#04x, manufacturer %d",
		l.HCIVersion, l.HCIRevision, l.LMPVersion, l.LMPSubversion, l.Manufacturer)
}

// ReadLocal asks the controller what it is, the ordinary way.
func ReadLocal(p *Port) (Local, error) {
	cmd, err := command(hciReadLocalVersion)
	if err != nil {
		return Local{}, err
	}
	if _, err := p.Write(cmd); err != nil {
		return Local{}, fmt.Errorf("ble: writing: %w", err)
	}

	said, err := waitReturn(p, hciReadLocalVersion, answers)
	if err != nil {
		return Local{}, err
	}
	if len(said) < 8 {
		return Local{}, fmt.Errorf("ble: a local version of %d bytes", len(said))
	}

	return Local{
		HCIVersion:    said[0],
		HCIRevision:   binary.LittleEndian.Uint16(said[1:]),
		LMPVersion:    said[3],
		Manufacturer:  binary.LittleEndian.Uint16(said[4:]),
		LMPSubversion: binary.LittleEndian.Uint16(said[6:]),
	}, nil
}

// Features is what the controller says it can do, as the eight LMP feature bytes.
type Features [8]byte

// Classic reports BR/EDR support, which is everything that is not Low Energy — a speaker, a
// headset, a phone connecting to play something.
func (f Features) Classic() bool { return f[4]&0x20 == 0 }

// LE reports Low Energy support, which is what a proxy scans with.
func (f Features) LE() bool { return f[4]&0x40 != 0 }

// EDR reports enhanced data rate, the 2 and 3 Mbps basic-rate modes.
func (f Features) EDR() bool { return f[3]&0x80 != 0 }

func (f Features) String() string {
	return fmt.Sprintf("classic %v, le %v, edr %v  [% x]", f.Classic(), f.LE(), f.EDR(), f[:])
}

// ReadFeatures asks the controller what it supports.
func ReadFeatures(p *Port) (Features, error) {
	var none Features

	cmd, err := command(hciReadLocalFeatures)
	if err != nil {
		return none, err
	}
	if _, err := p.Write(cmd); err != nil {
		return none, fmt.Errorf("ble: writing: %w", err)
	}

	said, err := waitReturn(p, hciReadLocalFeatures, answers)
	if err != nil {
		return none, err
	}
	if len(said) < 8 {
		return none, fmt.Errorf("ble: %d feature bytes, want 8", len(said))
	}
	return Features(said[:8]), nil
}

// Reset restarts the controller.
func Reset(p *Port) error {
	cmd, err := command(hciReset)
	if err != nil {
		return err
	}
	if _, err := p.Write(cmd); err != nil {
		return fmt.Errorf("resetting: %w", err)
	}
	if _, err := waitReturn(p, hciReset, answers); err != nil {
		return fmt.Errorf("resetting: %w", err)
	}
	p.Settle(switched)
	return nil
}

// waitReturn reads past whatever else is on the line for an opcode's command complete. What it
// skipped goes in the error: a silent line and a line full of other answers are different faults.
func waitReturn(p *Port, opcode uint16, within time.Duration) ([]byte, error) {
	deadline := time.Now().Add(within)

	var saw []string
	for {
		left := time.Until(deadline)
		if left <= 0 {
			if len(saw) == 0 {
				return nil, fmt.Errorf("ble: %#04x was not answered in %v", opcode, within)
			}
			return nil, fmt.Errorf("ble: %#04x was not answered in %v, but the chip said %s",
				opcode, within, strings.Join(saw, ", "))
		}

		e, err := p.Read(left)
		if err != nil {
			return nil, err
		}
		if e.Code == eventCommandComplete && len(e.Params) >= 3 &&
			binary.LittleEndian.Uint16(e.Params[1:]) == opcode {
			return complete(e, opcode)
		}
		saw = append(saw, fmt.Sprintf("%#02x [% x]", e.Code, e.Params))
	}
}
