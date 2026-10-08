package update

import (
	"fmt"
	"syscall"
)

const selinuxAttr = "security.selinux"

// init computes a service's domain from its binary's label and refuses one with no transition.
func copyLabel(from, to string) error {
	label := make([]byte, 256)

	n, err := syscall.Getxattr(from, selinuxAttr, label)
	if err != nil {
		return fmt.Errorf("update: reading the label of %s: %w", from, err)
	}
	if err := syscall.Setxattr(to, selinuxAttr, label[:n], 0); err != nil {
		return fmt.Errorf("update: labelling %s: %w", to, err)
	}
	return nil
}
