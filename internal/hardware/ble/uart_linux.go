//go:build linux && (board_checkers || board_cronos || board_rook)

package ble

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// Port is the controller's HCI node.
type Port struct {
	f *os.File

	// What has been read and not yet handed over: a read for one thing turns up another's.
	stream

	// wmu makes one write one packet.
	wmu sync.Mutex
}

// OpenNode takes an HCI character device, which speaks H4 with no line to configure.
func OpenNode(path string) (*Port, error) {
	// A blocking open of /dev/stpbt never returns: the driver powers the chip on inside it.
	f, err := os.OpenFile(path, unix.O_RDWR|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("ble: opening %s: %w", path, err)
	}
	return &Port{f: f}, nil
}

// Flush throws away whatever has been read and not handed over, for starting again after something
// went wrong mid-exchange.
func (p *Port) Flush() error {
	p.held = nil
	return nil
}

func (p *Port) Close() error { return p.f.Close() }

// readSome reads what has arrived. /dev/stpbt answers an empty read with zero bytes rather than
// EAGAIN, which os.File reports as EOF.
func (p *Port) readSome(buf []byte) (int, error) {
	n, err := p.f.Read(buf)
	if n == 0 && errors.Is(err, io.EOF) {
		time.Sleep(time.Millisecond)
		return 0, unix.EAGAIN
	}
	return n, err
}

// Write puts a packet on the line.
func (p *Port) Write(b []byte) (int, error) {
	p.wmu.Lock()
	defer p.wmu.Unlock()

	if err := p.f.SetWriteDeadline(time.Now().Add(answers)); err != nil {
		return 0, fmt.Errorf("ble: %w", err)
	}
	return p.f.Write(b)
}

// Send is Write, for a caller whose reader is a goroutine of its own.
func (p *Port) Send(b []byte) (int, error) { return p.Write(b) }

// Settle reads without wanting anything, so what the chip announces is heard rather than left for
// the next read to trip over.
func (p *Port) Settle(within time.Duration) {
	deadline := time.Now().Add(within)
	buf := make([]byte, 512)

	for time.Now().Before(deadline) {
		if err := p.f.SetReadDeadline(deadline); err != nil {
			return
		}

		n, err := p.readSome(buf)
		if n > 0 {
			p.held = append(p.held, buf[:n]...)
			continue
		}
		if err != nil && !errors.Is(err, unix.EAGAIN) {
			return
		}
	}
}

// Ask writes a command and reads until the chip answers or the time is up.
func (p *Port) Ask(cmd []byte, within time.Duration) (event, error) {
	if _, err := p.Write(cmd); err != nil {
		return event{}, fmt.Errorf("ble: writing: %w", err)
	}
	return p.Read(within)
}

// Raw reads bytes off the line and gives back whatever arrived, parsed or not, for finding out
// what the chip says rather than checking it said the expected thing.
func (p *Port) Raw(within time.Duration) ([]byte, error) {
	deadline := time.Now().Add(within)

	got := p.held
	p.held = nil

	buf := make([]byte, 512)
	for time.Now().Before(deadline) {
		if err := p.f.SetReadDeadline(deadline); err != nil {
			return got, fmt.Errorf("ble: %w", err)
		}

		n, err := p.readSome(buf)
		if n > 0 {
			got = append(got, buf[:n]...)
			continue
		}
		if errors.Is(err, os.ErrDeadlineExceeded) {
			break
		}
		if err != nil && !errors.Is(err, unix.EAGAIN) {
			return got, fmt.Errorf("ble: reading: %w", err)
		}
	}
	return got, nil
}

// Read waits for one event, keeping what arrives on the port: a read splits packets where it
// likes, and some commands are answered twice, so bytes past this event belong to the next read.
func (p *Port) Read(within time.Duration) (event, error) { return p.read(within) }

// Sniff reads the line for a while and reports every packet that arrived.
//
// For a command that timed out: the difference between a chip that said nothing and one whose
// answer nobody was waiting for.
func (p *Port) Sniff(within time.Duration) []string {
	deadline := time.Now().Add(within)
	buf := make([]byte, 512)

	for time.Now().Before(deadline) {
		if err := p.f.SetReadDeadline(deadline); err != nil {
			break
		}

		n, err := p.readSome(buf)
		if n > 0 {
			p.held = append(p.held, buf[:n]...)
		} else if err != nil && !errors.Is(err, unix.EAGAIN) && !errors.Is(err, os.ErrDeadlineExceeded) {
			break
		}

		p.drain()
	}

	var out []string
	for {
		e, ok := p.take()
		if !ok {
			break
		}
		out = append(out, fmt.Sprintf("event %#02x  % x", e.Code, e.Params))
	}

	if len(p.held) > 0 {
		out = append(out, fmt.Sprintf("unparsed: % x", p.held))
	}
	return out
}

func (p *Port) read(within time.Duration) (event, error) {
	var e event
	var ok bool

	if err := p.pump(within, func() bool { e, ok = p.take(); return ok }); err != nil {
		return event{}, err
	}
	return e, nil
}

// readACL is the same for data off a link.
//
// Its own queue, because a command's answer arriving between two packets of audio must not cost
// either of them: whichever the caller was not waiting for waits its turn instead of being dropped.
func (p *Port) readACL(within time.Duration) (aclPacket, error) {
	var a aclPacket
	var ok bool

	if err := p.pump(within, func() bool { a, ok = p.takeACL(); return ok }); err != nil {
		return aclPacket{}, err
	}
	return a, nil
}

// next is the oldest of either kind, for a reader handling both.
//
// Events before data when both are waiting. A command's answer decides what happens next and audio
// is already buffered, so the answer is the one worth a few microseconds.
func (p *Port) next(within time.Duration) (packet, error) {
	var got packet

	err := p.pump(within, func() bool {
		if e, ok := p.take(); ok {
			got = packet{kind: typeEvent, event: e}
			return true
		}
		if a, ok := p.takeACL(); ok {
			got = packet{kind: typeACL, acl: a}
			return true
		}
		return false
	})
	if err != nil {
		return packet{}, err
	}
	return got, nil
}

// pump reads the line until got says what it came for has arrived, or the time is up.
func (p *Port) pump(within time.Duration, got func() bool) error {
	deadline := time.Now().Add(within)
	buf := make([]byte, 512)

	for {
		// Anything already put aside for this reader, which may be the answer before the line is
		// touched at all.
		if got() {
			return nil
		}

		// Whatever has been read and not yet parsed into packets.
		if p.drain() {
			continue
		}

		if !time.Now().Before(deadline) {
			return fmt.Errorf("%w in %v", errQuiet, within)
		}

		// The line is opened non-blocking, so a read waits in the runtime poller rather than
		// returning. Without this it waits for as long as the chip stays quiet, which is forever.
		if err := p.f.SetReadDeadline(deadline); err != nil {
			return fmt.Errorf("ble: %w", err)
		}

		switch n, err := p.readSome(buf); {
		case n > 0:
			p.held = append(p.held, buf[:n]...)
		case errors.Is(err, os.ErrDeadlineExceeded):
			return fmt.Errorf("%w in %v", errQuiet, within)
		case errors.Is(err, unix.EAGAIN), err == nil:
			// Nothing on the line yet, which is most of the wait.
		default:
			return fmt.Errorf("ble: reading: %w", err)
		}
	}
}
