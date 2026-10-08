package prop

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"
)

const Socket = "/dev/socket/property_service"

const getprop = "/system/bin/getprop"

// bionic's prop_msg on Android 5.1; Android 8 replaced it with a length-prefixed protocol.
const (
	cmdSetProp = 1
	nameMax    = 32
	valueMax   = 92
	msgLen     = 4 + nameMax + valueMax
)

// init reports no error for a write its permission checks reject.
func Set(name, value string) error {
	if len(name) >= nameMax {
		return fmt.Errorf("prop: name %q exceeds %d bytes", name, nameMax-1)
	}
	if len(value) >= valueMax {
		return fmt.Errorf("prop: value for %q exceeds %d bytes", name, valueMax-1)
	}

	c, err := net.DialTimeout("unix", Socket, 2*time.Second)
	if err != nil {
		return fmt.Errorf("prop: dial: %w", err)
	}
	defer c.Close()

	msg := make([]byte, msgLen)
	binary.LittleEndian.PutUint32(msg[:4], cmdSetProp)
	copy(msg[4:4+nameMax-1], name)
	copy(msg[4+nameMax:msgLen-1], value)

	if err := c.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return fmt.Errorf("prop: deadline: %w", err)
	}
	if _, err := c.Write(msg); err != nil {
		return fmt.Errorf("prop: write: %w", err)
	}

	// init closes the connection once it has handled the message.
	if _, err := io.Copy(io.Discard, c); err != nil && !errors.Is(err, io.EOF) {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return nil
		}
		return fmt.Errorf("prop: wait: %w", err)
	}
	return nil
}

func Get(name string) (string, error) {
	out, err := exec.Command(getprop, name).Output()
	if err != nil {
		return "", fmt.Errorf("prop: reading %s: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func Stop(service string) error { return Set("ctl.stop", service) }

func Start(service string) error { return Set("ctl.start", service) }

func Restart(service string) error { return Set("ctl.restart", service) }
