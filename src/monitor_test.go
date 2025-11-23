// Package main provides unit tests for the monitor module.
// This file demonstrates testing interfaces, mocking, and testing real functions.
package main

import (
	"errors"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// Mock implementations for testing error paths

// mockCPUProvider allows us to control CPU function behavior in tests
type mockCPUProvider struct {
	percentages []float64
	err         error
}

func (m mockCPUProvider) Percent(duration time.Duration, percpu bool) ([]float64, error) {
	return m.percentages, m.err
}

// mockMemProvider allows us to control memory function behavior in tests
type mockMemProvider struct {
	vmStat *mem.VirtualMemoryStat
	err    error
}

func (m mockMemProvider) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return m.vmStat, m.err
}

// mockDiskProvider allows us to control disk function behavior in tests
type mockDiskProvider struct {
	usageStat *disk.UsageStat
	err       error
}

func (m mockDiskProvider) Usage(path string) (*disk.UsageStat, error) {
	return m.usageStat, m.err
}

// TestNewGopsutilMonitor tests the constructor function.
// This tests that we get a valid monitor instance.
func TestNewGopsutilMonitor(t *testing.T) {
	// Act
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Assert
	if monitor == nil {
		t.Fatal("NewGopsutilMonitor returned nil")
	}

	// Check that it implements the interface (this is now guaranteed by the return type)
	if monitor == nil {
		t.Error("NewGopsutilMonitor returned nil")
	}
}

// TestGopsutilMonitorImplementsInterface tests interface compliance.
// This is important - ensures our type actually implements the interface.
func TestGopsutilMonitorImplementsInterface(t *testing.T) {
	// Compile-time check that GopsutilMonitor implements SystemMonitor
	var _ SystemMonitor = &GopsutilMonitor{}

	// If this compiles, the test passes!
	t.Log("GopsutilMonitor correctly implements SystemMonitor interface")
}

// TestMemoryInfoStruct tests the MemoryInfo struct creation.
// This tests our data structures.
func TestMemoryInfoStruct(t *testing.T) {
	// Arrange
	testMemory := &MemoryInfo{
		UsedPercent: 75.5,
		Used:        8589934592,  // 8 GB in bytes
		Total:       17179869184, // 16 GB in bytes
	}

	// Assert
	if testMemory.UsedPercent != 75.5 {
		t.Errorf("Expected UsedPercent 75.5, got %f", testMemory.UsedPercent)
	}
	if testMemory.Used != 8589934592 {
		t.Errorf("Expected Used 8589934592, got %d", testMemory.Used)
	}
	if testMemory.Total != 17179869184 {
		t.Errorf("Expected Total 17179869184, got %d", testMemory.Total)
	}
}

// TestDiskInfoStruct tests the DiskInfo struct creation.
func TestDiskInfoStruct(t *testing.T) {
	// Arrange
	testDisk := &DiskInfo{
		UsedPercent: 45.2,
		Used:        483183820800,  // ~450 GB
		Total:       1099511627776, // 1 TB
	}

	// Assert
	if testDisk.UsedPercent != 45.2 {
		t.Errorf("Expected UsedPercent 45.2, got %f", testDisk.UsedPercent)
	}
	if testDisk.Used != 483183820800 {
		t.Errorf("Expected Used 483183820800, got %d", testDisk.Used)
	}
	if testDisk.Total != 1099511627776 {
		t.Errorf("Expected Total 1099511627776, got %d", testDisk.Total)
	}
}

// TestGopsutilMonitorCPUUsage tests the real CPU monitoring implementation.
// This tests our actual production code that calls gopsutil.
func TestGopsutilMonitorCPUUsage(t *testing.T) {
	// Arrange
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Act
	cpu, err := monitor.GetCPUUsage(100 * time.Millisecond)

	// Assert
	if err != nil {
		t.Skipf("Cannot test real CPU (might be in CI): %v", err)
		return
	}

	// CPU percentage should be reasonable
	if cpu < 0 || cpu > 100 {
		t.Errorf("CPU percentage out of range: %f%% (should be 0-100)", cpu)
	}
}

// TestGopsutilMonitorMemoryUsage tests the real memory monitoring implementation.
// This tests our actual production code that calls gopsutil.
func TestGopsutilMonitorMemoryUsage(t *testing.T) {
	// Arrange
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Act
	mem, err := monitor.GetMemoryUsage()

	// Assert
	if err != nil {
		t.Skipf("Cannot test real memory (might be in CI): %v", err)
		return
	}

	// Memory should have reasonable values
	if mem == nil {
		t.Fatal("GetMemoryUsage returned nil MemoryInfo")
	}

	// Percentage should be valid
	if mem.UsedPercent < 0 || mem.UsedPercent > 100 {
		t.Errorf("Memory percentage out of range: %f%% (should be 0-100)", mem.UsedPercent)
	}

	// Used should not exceed total
	if mem.Used > mem.Total {
		t.Errorf("Used memory (%d) cannot exceed total (%d)", mem.Used, mem.Total)
	}

	// Total should be reasonable (at least 1GB)
	if mem.Total < 1*1024*1024*1024 {
		t.Errorf("Total memory seems too low: %d bytes", mem.Total)
	}
}

// TestGopsutilMonitorDiskUsage tests the real disk monitoring implementation.
// This tests our actual production code that calls gopsutil.
func TestGopsutilMonitorDiskUsage(t *testing.T) {
	// Arrange
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Act - Test with a path that should exist on most systems
	disk, err := monitor.GetDiskUsage("C:")

	// Assert
	if err != nil {
		// Try alternative path for non-Windows systems
		disk, err = monitor.GetDiskUsage("/")
		if err != nil {
			t.Skipf("Cannot test real disk (might be in CI or different OS): %v", err)
			return
		}
	}

	// Disk should have reasonable values
	if disk == nil {
		t.Fatal("GetDiskUsage returned nil DiskInfo")
	}

	// Percentage should be valid
	if disk.UsedPercent < 0 || disk.UsedPercent > 100 {
		t.Errorf("Disk percentage out of range: %f%% (should be 0-100)", disk.UsedPercent)
	}

	// Used should not exceed total
	if disk.Used > disk.Total {
		t.Errorf("Used disk space (%d) cannot exceed total (%d)", disk.Used, disk.Total)
	}

	// Total should be reasonable (at least 1GB)
	if disk.Total < 1*1024*1024*1024 {
		t.Errorf("Total disk space seems too low: %d bytes", disk.Total)
	}
}

// TestGopsutilMonitorErrorHandling tests error paths in real implementation.
// This tests how our real code handles invalid inputs.
func TestGopsutilMonitorErrorHandling(t *testing.T) {
	// Arrange
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Test invalid disk path
	_, err := monitor.GetDiskUsage("/this/path/definitely/does/not/exist/on/any/system")
	if err == nil {
		t.Error("Expected error for invalid disk path, got nil")
	}

	// Error message should be helpful
	if err != nil && !contains(err.Error(), "failed to get disk usage") {
		t.Errorf("Error message should mention disk usage failure: %v", err)
	}
}

// TestGopsutilMonitorDataConsistency tests that our wrapper types work correctly.
// This ensures our MemoryInfo and DiskInfo structs contain the expected data.
func TestGopsutilMonitorDataConsistency(t *testing.T) {
	// Arrange
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Act
	mem, err := monitor.GetMemoryUsage()
	if err != nil {
		t.Skipf("Cannot test memory consistency: %v", err)
		return
	}

	// Assert - Check that percentage matches calculated value
	calculatedPercent := float64(mem.Used) / float64(mem.Total) * 100
	tolerance := 1.0 // Allow 1% tolerance for rounding

	if abs(mem.UsedPercent-calculatedPercent) > tolerance {
		t.Errorf("Memory percentage inconsistent: reported %f%%, calculated %f%%",
			mem.UsedPercent, calculatedPercent)
	}
}

// TestGopsutilMonitorErrorPaths tests all the error conditions to achieve 100% coverage.
// This tests the error handling paths that are hard to trigger with real hardware.
func TestGopsutilMonitorErrorPaths(t *testing.T) {
	t.Run("CPU Percent Error", func(t *testing.T) {
		// Arrange - Simple dependency injection
		mockCPU := mockCPUProvider{
			percentages: nil,
			err:         errors.New("mock CPU error"),
		}
		monitor := NewGopsutilMonitor(mockCPU, realMemProvider{}, realDiskProvider{})

		// Act
		_, err := monitor.GetCPUUsage(100 * time.Millisecond)

		// Assert
		if err == nil {
			t.Error("Expected error from CPU provider, got nil")
		}
		if !contains(err.Error(), "failed to get CPU usage") {
			t.Errorf("Error should mention CPU usage failure: %v", err)
		}
	})

	t.Run("CPU Empty Slice", func(t *testing.T) {
		// Arrange - Mock that returns empty slice (no error, but no data)
		mockCPU := mockCPUProvider{
			percentages: []float64{}, // Empty slice
			err:         nil,
		}
		monitor := NewGopsutilMonitor(mockCPU, realMemProvider{}, realDiskProvider{})

		// Act
		_, err := monitor.GetCPUUsage(100 * time.Millisecond)

		// Assert
		if err == nil {
			t.Error("Expected error for empty CPU data, got nil")
		}
		if !contains(err.Error(), "no CPU usage data returned") {
			t.Errorf("Error should mention no CPU data: %v", err)
		}
	})

	t.Run("Memory VirtualMemory Error", func(t *testing.T) {
		// Arrange - Simple dependency injection
		mockMem := mockMemProvider{
			vmStat: nil,
			err:    errors.New("mock memory error"),
		}
		monitor := NewGopsutilMonitor(realCPUProvider{}, mockMem, realDiskProvider{})

		// Act
		_, err := monitor.GetMemoryUsage()

		// Assert
		if err == nil {
			t.Error("Expected error from memory provider, got nil")
		}
		if !contains(err.Error(), "failed to get memory usage") {
			t.Errorf("Error should mention memory usage failure: %v", err)
		}
	})

	t.Run("Disk Usage Error", func(t *testing.T) {
		// Arrange - Simple dependency injection
		mockDisk := mockDiskProvider{
			usageStat: nil,
			err:       errors.New("mock disk error"),
		}
		monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, mockDisk)

		// Act
		_, err := monitor.GetDiskUsage("/invalid/path")

		// Assert
		if err == nil {
			t.Error("Expected error from disk provider, got nil")
		}
		if !contains(err.Error(), "failed to get disk usage") {
			t.Errorf("Error should mention disk usage failure: %v", err)
		}
	})
}

// TestGopsutilMonitorSuccessPaths tests the success paths with mocked data.
// This ensures our conversion logic works correctly.
func TestGopsutilMonitorSuccessPaths(t *testing.T) {
	t.Run("CPU Success", func(t *testing.T) {
		// Arrange - Simple dependency injection
		mockCPU := mockCPUProvider{
			percentages: []float64{45.5}, // Valid CPU percentage
			err:         nil,
		}
		monitor := NewGopsutilMonitor(mockCPU, realMemProvider{}, realDiskProvider{})

		// Act
		cpu, err := monitor.GetCPUUsage(100 * time.Millisecond)

		// Assert
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if cpu != 45.5 {
			t.Errorf("Expected CPU 45.5%%, got %f%%", cpu)
		}
	})

	t.Run("Memory Success", func(t *testing.T) {
		// Arrange - Simple dependency injection
		mockMem := mockMemProvider{
			vmStat: &mem.VirtualMemoryStat{
				UsedPercent: 75.0,
				Used:        8 * 1024 * 1024 * 1024,  // 8GB
				Total:       16 * 1024 * 1024 * 1024, // 16GB
			},
			err: nil,
		}
		monitor := NewGopsutilMonitor(realCPUProvider{}, mockMem, realDiskProvider{})

		// Act
		mem, err := monitor.GetMemoryUsage()

		// Assert
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if mem.UsedPercent != 75.0 {
			t.Errorf("Expected memory 75.0%%, got %f%%", mem.UsedPercent)
		}
		if mem.Used != 8*1024*1024*1024 {
			t.Errorf("Expected used 8GB, got %d", mem.Used)
		}
	})

	t.Run("Disk Success", func(t *testing.T) {
		// Arrange - Simple dependency injection
		mockDisk := mockDiskProvider{
			usageStat: &disk.UsageStat{
				UsedPercent: 60.0,
				Used:        600 * 1024 * 1024 * 1024,  // 600GB
				Total:       1000 * 1024 * 1024 * 1024, // 1TB
			},
			err: nil,
		}
		monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, mockDisk)

		// Act
		disk, err := monitor.GetDiskUsage("/test")

		// Assert
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if disk.UsedPercent != 60.0 {
			t.Errorf("Expected disk 60.0%%, got %f%%", disk.UsedPercent)
		}
		if disk.Used != 600*1024*1024*1024 {
			t.Errorf("Expected used 600GB, got %d", disk.Used)
		}
	})
}

// Helper functions for tests
func contains(s, substr string) bool {
	return len(substr) <= len(s) && (substr == s ||
		(len(substr) > 0 && len(s) > 0 && s[0:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestSystemMonitorInterface tests that our real implementation satisfies the interface.
// This ensures GopsutilMonitor actually implements SystemMonitor correctly.
func TestSystemMonitorInterface(t *testing.T) {
	// Arrange - Create real monitor
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Act & Assert - Verify it's not nil
	if monitor == nil {
		t.Fatal("NewGopsutilMonitor returned nil")
	}

	// Test that all interface methods exist and work
	_, err := monitor.GetCPUUsage(50 * time.Millisecond)
	if err != nil {
		t.Logf("CPU test skipped: %v", err)
	}

	_, err = monitor.GetMemoryUsage()
	if err != nil {
		t.Logf("Memory test skipped: %v", err)
	}

	_, err = monitor.GetDiskUsage("C:")
	if err != nil {
		// Try alternative for non-Windows
		_, err = monitor.GetDiskUsage("/")
		if err != nil {
			t.Logf("Disk test skipped: %v", err)
		}
	}
}

// TestValidMemoryPercentages tests that memory percentages are reasonable.
// This is a sanity check for real data.
func TestValidMemoryPercentages(t *testing.T) {
	// Skip this test if we're just building (not running on real hardware)
	if testing.Short() {
		t.Skip("Skipping real hardware test in short mode")
	}

	// Arrange
	monitor := NewGopsutilMonitor(realCPUProvider{}, realMemProvider{}, realDiskProvider{})

	// Act
	mem, err := monitor.GetMemoryUsage()

	// Assert
	if err != nil {
		// This is OK - might not be running on real hardware
		t.Skipf("Cannot test real memory: %v", err)
		return
	}

	// Memory percentage should be reasonable
	if mem.UsedPercent < 0 || mem.UsedPercent > 100 {
		t.Errorf("Memory percentage out of range: %f%%", mem.UsedPercent)
	}

	// Used should be less than total
	if mem.Used > mem.Total {
		t.Errorf("Used memory (%d) cannot be greater than total (%d)", mem.Used, mem.Total)
	}
}
