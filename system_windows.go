//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func getDiskUsage() string {
	out, err := exec.Command("wmic", "logicaldisk", "where", "DeviceID='C:'", "get", "FreeSpace,Size", "/format:list").CombinedOutput()
	if err != nil {
		return "N/A"
	}
	var free, total uint64
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FreeSpace=") {
			fmt.Sscanf(strings.TrimPrefix(line, "FreeSpace="), "%d", &free)
		}
		if strings.HasPrefix(line, "Size=") {
			fmt.Sscanf(strings.TrimPrefix(line, "Size="), "%d", &total)
		}
	}
	if total == 0 {
		return "N/A"
	}
	freeGB := float64(free) / (1024 * 1024 * 1024)
	totalGB := float64(total) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.1fGB free / %.1fGB total", freeGB, totalGB)
}

func getRAMUsage() string {
	out, err := exec.Command("wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize", "/format:list").CombinedOutput()
	if err != nil {
		return "N/A"
	}
	var freeKB, totalKB uint64
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FreePhysicalMemory=") {
			fmt.Sscanf(strings.TrimPrefix(line, "FreePhysicalMemory="), "%d", &freeKB)
		}
		if strings.HasPrefix(line, "TotalVisibleMemorySize=") {
			fmt.Sscanf(strings.TrimPrefix(line, "TotalVisibleMemorySize="), "%d", &totalKB)
		}
	}
	if totalKB == 0 {
		return "N/A"
	}
	totalMB := float64(totalKB) / 1024
	availMB := float64(freeKB) / 1024
	if totalMB > 1024 {
		return fmt.Sprintf("%.1fGB available / %.1fGB total", availMB/1024, totalMB/1024)
	}
	return fmt.Sprintf("%.0fMB available / %.0fMB total", availMB, totalMB)
}
