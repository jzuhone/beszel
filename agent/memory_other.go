//go:build !illumos

package agent

import (
	"errors"

	"github.com/shirou/gopsutil/v4/mem"
)

func illumosVirtualMemory() (*mem.VirtualMemoryStat, error) {
	return nil, errors.ErrUnsupported
}
