//go:build !linux && !freebsd && !illumos && !solaris

package zfs

import "errors"

func ARCSize() (uint64, error) {
	return 0, errors.ErrUnsupported
}
