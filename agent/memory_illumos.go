//go:build illumos

package agent

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/mem"
)

// illumosVirtualMemory returns virtual memory statistics for Illumos by
// querying kstat for system page counters.
//
// Fields populated:
//   - Total      = physmem * pagesize (total physical RAM)
//   - Free       = freemem * pagesize
//   - Available  = freemem * pagesize
//   - Used       = Total - Free  (all non-free memory, including kernel)
//   - Cached     = pp_kernel * pagesize  (kernel memory; shown as "cache/buffers")
//   - UsedPercent derived from Used/Total
//
// This mirrors the Linux approach where v.Cached includes ZFS ARC — the ZFS
// subtraction logic in system.go then correctly moves arcSize out of v.Used
// into its own MemZfsArc category.
func illumosVirtualMemory() (*mem.VirtualMemoryStat, error) {
	pagesizeOut, err := exec.Command("pagesize").Output()
	if err != nil {
		return nil, fmt.Errorf("pagesize: %w", err)
	}
	pageSize, err := strconv.ParseUint(strings.TrimSpace(string(pagesizeOut)), 10, 64)
	if err != nil || pageSize == 0 {
		return nil, fmt.Errorf("invalid page size: %q", pagesizeOut)
	}

	out, err := exec.Command("kstat", "-p",
		"unix:0:system_pages:physmem",
		"unix:0:system_pages:freemem",
		"unix:0:system_pages:pp_kernel",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("kstat system_pages: %w", err)
	}

	pages := parseKstatPages(string(out))
	physmem := pages["unix:0:system_pages:physmem"]
	freemem := pages["unix:0:system_pages:freemem"]
	ppKernel := pages["unix:0:system_pages:pp_kernel"]

	if physmem == 0 {
		return nil, fmt.Errorf("physmem is zero in kstat output")
	}

	total := physmem * pageSize
	free := freemem * pageSize
	used := total - free

	var usedPercent float64
	if total > 0 {
		usedPercent = float64(used) / float64(total) * 100.0
	}

	return &mem.VirtualMemoryStat{
		Total:       total,
		Free:        free,
		Available:   free,
		Used:        used,
		UsedPercent: usedPercent,
		Cached:      ppKernel * pageSize,
	}, nil
}

// parseKstatPages parses "kstat -p" output into a map of statistic name → value.
// Each line has the form: "module:instance:name:statistic\tvalue"
func parseKstatPages(output string) map[string]uint64 {
	result := make(map[string]uint64)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, err := strconv.ParseUint(fields[len(fields)-1], 10, 64)
		if err != nil {
			continue
		}
		result[fields[0]] = val
	}
	return result
}
