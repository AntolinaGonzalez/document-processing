package models

import (
	"testing"
	"time"
)

func TestProcess_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		state    ProcessState
		expected bool
	}{
		{"pending should be active", StatePending, true},
		{"running should be active", StateRunning, true},
		{"completed should not be active", StateCompleted, false},
		{"failed should not be active", StateFailed, false},
		{"stopped should not be active", StateStopped, false},
		{"paused should not be active", StatePaused, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Process{State: tt.state}
			if got := p.IsActive(); got != tt.expected {
				t.Errorf("Process.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProcess_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		state    ProcessState
		expected bool
	}{
		{"pending should not be completed", StatePending, false},
		{"running should not be completed", StateRunning, false},
		{"completed should be completed", StateCompleted, true},
		{"failed should be completed", StateFailed, true},
		{"stopped should be completed", StateStopped, true},
		{"paused should not be completed", StatePaused, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Process{State: tt.state}
			if got := p.IsCompleted(); got != tt.expected {
				t.Errorf("Process.IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProcess_CalculateProgress(t *testing.T) {
	tests := []struct {
		name           string
		totalFiles     int
		processedFiles int
		expected       float64
	}{
		{"zero total files", 0, 0, 0.0},
		{"zero processed files", 10, 0, 0.0},
		{"half progress", 10, 5, 50.0},
		{"full progress", 10, 10, 100.0},
		{"over progress should cap at 100", 10, 15, 100.0},
		{"partial progress", 7, 3, 42.857142857142854},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Process{
				TotalFiles:     tt.totalFiles,
				ProcessedFiles: tt.processedFiles,
			}
			if got := p.CalculateProgress(); got != tt.expected {
				t.Errorf("Process.CalculateProgress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProcessState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    ProcessState
		expected string
	}{
		{"pending", StatePending, "PENDING"},
		{"running", StateRunning, "RUNNING"},
		{"paused", StatePaused, "PAUSED"},
		{"completed", StateCompleted, "COMPLETED"},
		{"failed", StateFailed, "FAILED"},
		{"stopped", StateStopped, "STOPPED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.state); got != tt.expected {
				t.Errorf("ProcessState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProcess_JSONSerialization(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC)
	errorMsg := "test error"

	p := &Process{
		ProcessID:      "test-uuid",
		FolderPath:     "/test/path",
		BatchSize:      10,
		State:          StateRunning,
		StartTime:      startTime,
		EndTime:        &endTime,
		LastError:      &errorMsg,
		TotalFiles:     100,
		ProcessedFiles: 50,
		Progress:       50.0,
		CreatedAt:      startTime,
		UpdatedAt:      endTime,
	}

	// Test that the struct can be marshaled to JSON
	_, err := p.MarshalJSON()
	if err != nil {
		t.Errorf("Process.MarshalJSON() error = %v", err)
	}
} 