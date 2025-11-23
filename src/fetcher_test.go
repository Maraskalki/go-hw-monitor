package main

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// MockSystemMonitor for testing purposes
type MockSystemMonitor struct {
	CPUUsage    float64
	CPUError    error
	MemoryInfo  *MemoryInfo
	MemoryError error
	DiskInfo    *DiskInfo
	DiskError   error
}

func (m *MockSystemMonitor) GetCPUUsage(duration time.Duration) (float64, error) {
	if m.CPUError != nil {
		return 0, m.CPUError
	}
	return m.CPUUsage, nil
}

func (m *MockSystemMonitor) GetMemoryUsage() (*MemoryInfo, error) {
	if m.MemoryError != nil {
		return nil, m.MemoryError
	}
	return m.MemoryInfo, nil
}

func (m *MockSystemMonitor) GetDiskUsage(path string) (*DiskInfo, error) {
	if m.DiskError != nil {
		return nil, m.DiskError
	}
	return m.DiskInfo, nil
}

func TestFetchSystemStats(t *testing.T) {
	// Test successful data collection
	t.Run("Success", func(t *testing.T) {
		mock := &MockSystemMonitor{
			CPUUsage: 75.5,
			MemoryInfo: &MemoryInfo{
				UsedPercent: 60.0,
				Used:        8 * 1024 * 1024 * 1024,  // 8GB
				Total:       16 * 1024 * 1024 * 1024, // 16GB
			},
			DiskInfo: &DiskInfo{
				UsedPercent: 45.0,
				Used:        450 * 1024 * 1024 * 1024,  // 450GB
				Total:       1000 * 1024 * 1024 * 1024, // 1TB
			},
		}

		statsCh := make(chan SystemStats, 1)
		fetchSystemStats(mock, statsCh)

		select {
		case stats := <-statsCh:
			if stats.CPUUsage != 75.5 {
				t.Errorf("Expected CPU usage 75.5, got %f", stats.CPUUsage)
			}
			if stats.MemoryUsage != 60.0 {
				t.Errorf("Expected memory usage 60.0%%, got %f%%", stats.MemoryUsage)
			}
			if stats.MemoryUsed != 8.0 {
				t.Errorf("Expected memory used 8.0GB, got %fGB", stats.MemoryUsed)
			}
			if stats.MemoryTotal != 16.0 {
				t.Errorf("Expected memory total 16.0GB, got %fGB", stats.MemoryTotal)
			}
			if stats.DiskUsage != 45.0 {
				t.Errorf("Expected disk usage 45.0%%, got %f%%", stats.DiskUsage)
			}
			if stats.DiskUsed != 450.0 {
				t.Errorf("Expected disk used 450.0GB, got %fGB", stats.DiskUsed)
			}
			if stats.DiskTotal != 1000.0 {
				t.Errorf("Expected disk total 1000.0GB, got %fGB", stats.DiskTotal)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for stats")
		}
	})

	// Test with errors (should continue with partial data)
	t.Run("WithErrors", func(t *testing.T) {
		mock := &MockSystemMonitor{
			CPUUsage:    75.5,
			CPUError:    nil,
			MemoryError: errors.New("memory error"),
			DiskError:   errors.New("disk error"),
		}

		statsCh := make(chan SystemStats, 1)
		fetchSystemStats(mock, statsCh)

		select {
		case stats := <-statsCh:
			// CPU should be populated
			if stats.CPUUsage != 75.5 {
				t.Errorf("Expected CPU usage 75.5, got %f", stats.CPUUsage)
			}
			// Memory and disk should be zero due to errors
			if stats.MemoryUsage != 0 {
				t.Errorf("Expected memory usage 0 due to error, got %f", stats.MemoryUsage)
			}
			if stats.DiskUsage != 0 {
				t.Errorf("Expected disk usage 0 due to error, got %f", stats.DiskUsage)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for stats")
		}
	})
}

func TestFetchCPUMetric(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &MockSystemMonitor{
			CPUUsage: 85.5,
		}

		var wg sync.WaitGroup
		results := make(chan MetricResult, 1)

		wg.Add(1)
		go fetchCPUMetric(mock, &wg, results)
		wg.Wait()
		close(results)

		result := <-results
		if result.Type != "cpu" {
			t.Errorf("Expected type 'cpu', got '%s'", result.Type)
		}
		if result.Error != nil {
			t.Errorf("Expected no error, got: %v", result.Error)
		}
		if cpuUsage, ok := result.Value.(float64); !ok {
			t.Error("Expected float64 value")
		} else if cpuUsage != 85.5 {
			t.Errorf("Expected CPU usage 85.5, got %f", cpuUsage)
		}
	})

	t.Run("Error", func(t *testing.T) {
		mock := &MockSystemMonitor{
			CPUError: errors.New("cpu error"),
		}

		var wg sync.WaitGroup
		results := make(chan MetricResult, 1)

		wg.Add(1)
		go fetchCPUMetric(mock, &wg, results)
		wg.Wait()
		close(results)

		result := <-results
		if result.Type != "cpu" {
			t.Errorf("Expected type 'cpu', got '%s'", result.Type)
		}
		if result.Error == nil {
			t.Error("Expected error, got nil")
		}
		if result.Value != nil {
			t.Errorf("Expected nil value on error, got %v", result.Value)
		}
	})
}

func TestFetchMemoryMetric(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedMemory := &MemoryInfo{
			UsedPercent: 70.0,
			Used:        7 * 1024 * 1024 * 1024,  // 7GB
			Total:       10 * 1024 * 1024 * 1024, // 10GB
		}

		mock := &MockSystemMonitor{
			MemoryInfo: expectedMemory,
		}

		var wg sync.WaitGroup
		results := make(chan MetricResult, 1)

		wg.Add(1)
		go fetchMemoryMetric(mock, &wg, results)
		wg.Wait()
		close(results)

		result := <-results
		if result.Type != "memory" {
			t.Errorf("Expected type 'memory', got '%s'", result.Type)
		}
		if result.Error != nil {
			t.Errorf("Expected no error, got: %v", result.Error)
		}
		if memInfo, ok := result.Value.(*MemoryInfo); !ok {
			t.Error("Expected *MemoryInfo value")
		} else {
			if memInfo.UsedPercent != 70.0 {
				t.Errorf("Expected memory used percent 70.0, got %f", memInfo.UsedPercent)
			}
			if memInfo.Used != 7*1024*1024*1024 {
				t.Errorf("Expected memory used %d, got %d", 7*1024*1024*1024, memInfo.Used)
			}
		}
	})

	t.Run("Error", func(t *testing.T) {
		mock := &MockSystemMonitor{
			MemoryError: errors.New("memory error"),
		}

		var wg sync.WaitGroup
		results := make(chan MetricResult, 1)

		wg.Add(1)
		go fetchMemoryMetric(mock, &wg, results)
		wg.Wait()
		close(results)

		result := <-results
		if result.Type != "memory" {
			t.Errorf("Expected type 'memory', got '%s'", result.Type)
		}
		if result.Error == nil {
			t.Error("Expected error, got nil")
		}
		if result.Value != nil {
			t.Errorf("Expected nil value on error, got %v", result.Value)
		}
	})
}

func TestFetchDiskMetric(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedDisk := &DiskInfo{
			UsedPercent: 80.0,
			Used:        800 * 1024 * 1024 * 1024,  // 800GB
			Total:       1000 * 1024 * 1024 * 1024, // 1TB
		}

		mock := &MockSystemMonitor{
			DiskInfo: expectedDisk,
		}

		var wg sync.WaitGroup
		results := make(chan MetricResult, 1)

		wg.Add(1)
		go fetchDiskMetric(mock, &wg, results)
		wg.Wait()
		close(results)

		result := <-results
		if result.Type != "disk" {
			t.Errorf("Expected type 'disk', got '%s'", result.Type)
		}
		if result.Error != nil {
			t.Errorf("Expected no error, got: %v", result.Error)
		}
		if diskInfo, ok := result.Value.(*DiskInfo); !ok {
			t.Error("Expected *DiskInfo value")
		} else {
			if diskInfo.UsedPercent != 80.0 {
				t.Errorf("Expected disk used percent 80.0, got %f", diskInfo.UsedPercent)
			}
			if diskInfo.Used != 800*1024*1024*1024 {
				t.Errorf("Expected disk used %d, got %d", 800*1024*1024*1024, diskInfo.Used)
			}
		}
	})

	t.Run("Error", func(t *testing.T) {
		mock := &MockSystemMonitor{
			DiskError: errors.New("disk error"),
		}

		var wg sync.WaitGroup
		results := make(chan MetricResult, 1)

		wg.Add(1)
		go fetchDiskMetric(mock, &wg, results)
		wg.Wait()
		close(results)

		result := <-results
		if result.Type != "disk" {
			t.Errorf("Expected type 'disk', got '%s'", result.Type)
		}
		if result.Error == nil {
			t.Error("Expected error, got nil")
		}
		if result.Value != nil {
			t.Errorf("Expected nil value on error, got %v", result.Value)
		}
	})
}

func TestSystemStats(t *testing.T) {
	// Test the SystemStats struct can be created and populated
	stats := SystemStats{
		CPUUsage:    50.0,
		MemoryUsage: 60.0,
		MemoryUsed:  8.0,
		MemoryTotal: 16.0,
		DiskUsage:   70.0,
		DiskUsed:    700.0,
		DiskTotal:   1000.0,
	}

	if stats.CPUUsage != 50.0 {
		t.Errorf("Expected CPU usage 50.0, got %f", stats.CPUUsage)
	}
	if stats.MemoryUsage != 60.0 {
		t.Errorf("Expected memory usage 60.0, got %f", stats.MemoryUsage)
	}
	if stats.DiskUsage != 70.0 {
		t.Errorf("Expected disk usage 70.0, got %f", stats.DiskUsage)
	}
}

func TestMetricResult(t *testing.T) {
	// Test successful result
	successResult := MetricResult{
		Type:  "cpu",
		Value: 75.5,
		Error: nil,
	}

	if successResult.Type != "cpu" {
		t.Errorf("Expected type 'cpu', got '%s'", successResult.Type)
	}
	if successResult.Value != 75.5 {
		t.Errorf("Expected value 75.5, got %v", successResult.Value)
	}
	if successResult.Error != nil {
		t.Errorf("Expected no error, got %v", successResult.Error)
	}

	// Test error result
	errorResult := MetricResult{
		Type:  "memory",
		Value: nil,
		Error: errors.New("test error"),
	}

	if errorResult.Type != "memory" {
		t.Errorf("Expected type 'memory', got '%s'", errorResult.Type)
	}
	if errorResult.Value != nil {
		t.Errorf("Expected nil value, got %v", errorResult.Value)
	}
	if errorResult.Error == nil {
		t.Error("Expected error, got nil")
	}
}
