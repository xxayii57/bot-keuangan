//go:build linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func getDiskUsage() string {
	var stat syscall.Statfs_t
	wd, err := os.Getwd()
	if err != nil {
		wd = "/"
	}
	err = syscall.Statfs(wd, &stat)
	if err != nil {
		return "N/A"
	}

	freeBytes := stat.Bavail * uint64(stat.Bsize)
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	freeGB := float64(freeBytes) / (1024 * 1024 * 1024)
	totalGB := float64(totalBytes) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.1fGB free / %.1fGB total", freeGB, totalGB)
}

func getRAMUsage() string {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return "N/A"
	}
	defer file.Close()

	var memTotal, memAvailable, memFree uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimSuffix(fields[0], ":")
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch name {
		case "MemTotal":
			memTotal = val
		case "MemAvailable":
			memAvailable = val
		case "MemFree":
			memFree = val
		}
	}

	if memAvailable == 0 {
		memAvailable = memFree
	}

	totalMB := float64(memTotal) / 1024
	availableMB := float64(memAvailable) / 1024

	if totalMB > 1024 {
		return fmt.Sprintf("%.1fGB available / %.1fGB total", availableMB/1024, totalMB/1024)
	}
	return fmt.Sprintf("%.0fMB available / %.0fMB total", availableMB, totalMB)
}
