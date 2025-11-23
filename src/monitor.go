// Package main provides hardware monitoring implementations.
// This file contains the SystemMonitor interface and its implementations.
package main

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemMonitor interface defines what we need from any monitoring system.
// This is the "contract" - any type that implements these methods can be used.
// Interfaces in Go make code flexible and testable.
type SystemMonitor interface {
	// GetCPUUsage returns CPU percentage (0-100) over the given duration
	GetCPUUsage(duration time.Duration) (float64, error)

	// GetMemoryUsage returns memory statistics
	GetMemoryUsage() (*MemoryInfo, error)

	// GetDiskUsage returns disk statistics for the given path
	GetDiskUsage(path string) (*DiskInfo, error)
}

// MemoryInfo holds clean memory statistics (wrapper around gopsutil data)
type MemoryInfo struct {
	UsedPercent float64 // Memory percentage (0-100)
	Used        uint64  // Memory used in bytes
	Total       uint64  // Total memory in bytes
}

// DiskInfo holds clean disk statistics (wrapper around gopsutil data)
type DiskInfo struct {
	UsedPercent float64 // Disk percentage (0-100)
	Used        uint64  // Disk used in bytes
	Total       uint64  // Total disk space in bytes
}

// GopsutilMonitor is our production implementation of SystemMonitor.
// It uses the gopsutil library to get real system metrics.
// This is called a "concrete type" that implements the interface.
type GopsutilMonitor struct {
	// Empty struct - we don't need to store any data
	// All the work is done by calling gopsutil functions
}

// NewGopsutilMonitor creates a new instance of the production monitor.
// This is a constructor function - a common Go pattern for creating instances.
func NewGopsutilMonitor() SystemMonitor {
	return &GopsutilMonitor{}
}

// GetCPUUsage implements SystemMonitor interface for CPU monitoring.
// This wraps the gopsutil cpu.Percent function in our clean interface.
func (g *GopsutilMonitor) GetCPUUsage(duration time.Duration) (float64, error) {
	// Same logic as before, but now it's in an interface method
	percentages, err := cpu.Percent(duration, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	if len(percentages) == 0 {
		return 0, fmt.Errorf("no CPU usage data returned")
	}

	return percentages[0], nil
}

// GetMemoryUsage implements SystemMonitor interface for memory monitoring.
// This wraps gopsutil mem.VirtualMemory in our clean interface.
func (g *GopsutilMonitor) GetMemoryUsage() (*MemoryInfo, error) {
	// Call gopsutil to get raw data
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}

	// Convert to our clean format
	return &MemoryInfo{
		UsedPercent: vmStat.UsedPercent,
		Used:        vmStat.Used,
		Total:       vmStat.Total,
	}, nil
}

// GetDiskUsage implements SystemMonitor interface for disk monitoring.
// This wraps gopsutil disk.Usage in our clean interface.
func (g *GopsutilMonitor) GetDiskUsage(path string) (*DiskInfo, error) {
	// Call gopsutil to get raw data
	diskStat, err := disk.Usage(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage for %s: %w", path, err)
	}

	// Convert to our clean format
	return &DiskInfo{
		UsedPercent: diskStat.UsedPercent,
		Used:        diskStat.Used,
		Total:       diskStat.Total,
	}, nil
}
