//go:build illumos || solaris

package zfs

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ARCSize returns the current ZFS ARC size in bytes by querying kstat.
func ARCSize() (uint64, error) {
	out, err := exec.Command("kstat", "-p", "zfs:0:arcstats:size").Output()
	if err != nil {
		return 0, fmt.Errorf("kstat: %w", err)
	}
	// Output format: "zfs:0:arcstats:size\t<value>"
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected kstat output: %s", out)
	}
	return strconv.ParseUint(fields[len(fields)-1], 10, 64)
}
