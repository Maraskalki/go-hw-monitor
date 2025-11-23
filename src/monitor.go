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

// Internal interfaces for dependency injection and testing
// These allow us to mock the gopsutil calls in tests

// cpuProvider wraps gopsutil cpu functions
type cpuProvider interface {
	Percent(duration time.Duration, percpu bool) ([]float64, error)
}

// memProvider wraps gopsutil memory functions
type memProvider interface {
	VirtualMemory() (*mem.VirtualMemoryStat, error)
}

// diskProvider wraps gopsutil disk functions
type diskProvider interface {
	Usage(path string) (*disk.UsageStat, error)
}

// Real implementations of the providers
type realCPUProvider struct{}
type realMemProvider struct{}
type realDiskProvider struct{}

func (r realCPUProvider) Percent(duration time.Duration, percpu bool) ([]float64, error) {
	return cpu.Percent(duration, percpu)
}

func (r realMemProvider) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (r realDiskProvider) Usage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}

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
	// Dependencies for testing - these can be mocked
	cpu  cpuProvider
	mem  memProvider
	disk diskProvider
}

// NewGopsutilMonitor creates a new monitor with injectable dependencies.
// For production use, pass real providers. For testing, pass mocks.
func NewGopsutilMonitor(cpuProv cpuProvider, memProv memProvider, diskProv diskProvider) SystemMonitor {
	return &GopsutilMonitor{
		cpu:  cpuProv,
		mem:  memProv,
		disk: diskProv,
	}
}

// GetCPUUsage implements SystemMonitor interface for CPU monitoring.
// This wraps the gopsutil cpu.Percent function in our clean interface.
func (g *GopsutilMonitor) GetCPUUsage(duration time.Duration) (float64, error) {
	// Use injected dependency instead of calling cpu.Percent directly
	percentages, err := g.cpu.Percent(duration, false)
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
	// Use injected dependency instead of calling mem.VirtualMemory directly
	vmStat, err := g.mem.VirtualMemory()
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
	// Use injected dependency instead of calling disk.Usage directly
	diskStat, err := g.disk.Usage(path)
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
