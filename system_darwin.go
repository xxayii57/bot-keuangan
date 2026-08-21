//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func getDiskUsage() string {
	out, err := exec.Command("df", "-h", ".").CombinedOutput()
	if err != nil {
		return "N/A"
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			return fmt.Sprintf("%s used / %s total (%s avail)", fields[2], fields[1], fields[3])
		}
	}
	return "N/A"
}

func getRAMUsage() string {
	out, err := exec.Command("sysctl", "hw.memsize").CombinedOutput()
	if err != nil {
		return "N/A"
	}
	var totalBytes uint64
	fmt.Sscanf(strings.TrimSpace(string(out)), "hw.memsize: %d", &totalBytes)

	out, err = exec.Command("vm_stat").CombinedOutput()
	if err != nil {
		return "N/A"
	}
	var freePages, speculativePages uint64
	var pageSize uint64 = 4096
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "page size of") {
			fmt.Sscanf(line, "page size of %d", &pageSize)
		}
		if strings.HasPrefix(line, "Pages free:") {
			fmt.Sscanf(strings.TrimPrefix(line, "Pages free:"), "%d", &freePages)
		}
		if strings.HasPrefix(line, "Pages speculative:") {
			fmt.Sscanf(strings.TrimPrefix(line, "Pages speculative:"), "%d", &speculativePages)
		}
	}
	freeBytes := (freePages + speculativePages) * pageSize
	totalMB := float64(totalBytes) / (1024 * 1024)
	availMB := float64(freeBytes) / (1024 * 1024)
	if totalMB > 1024 {
		return fmt.Sprintf("%.1fGB available / %.1fGB total", availMB/1024, totalMB/1024)
	}
	return fmt.Sprintf("%.0fMB available / %.0fMB total", availMB, totalMB)
}
