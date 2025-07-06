package models

import (
	"time"
)

type ProcessState string

const (
	StatePending   ProcessState = "PENDING"
	StateRunning   ProcessState = "RUNNING"
	StatePaused    ProcessState = "PAUSED"
	StateCompleted ProcessState = "COMPLETED"
	StateFailed    ProcessState = "FAILED"
	StateStopped   ProcessState = "STOPPED"
)

type Process struct {
	ProcessID      string       `json:"process_id"`
	FolderPath     string       `json:"folder_path"`
	BatchSize      int          `json:"batch_size"`
	State          ProcessState `json:"state"`
	StartTime      time.Time    `json:"start_time"`
	EndTime        *time.Time   `json:"end_time,omitempty"`
	LastError      *string      `json:"last_error,omitempty"`
	TotalFiles     int          `json:"total_files,omitempty"`
	ProcessedFiles int          `json:"processed_files,omitempty"`
	Progress       float64      `json:"progress,omitempty"`
	CreatedAt      time.Time    `json:"created_at,omitempty"`
	UpdatedAt      time.Time    `json:"updated_at,omitempty"`
}

// IsActive returns true if the process is currently running or pending
func (p *Process) IsActive() bool {
	return p.State == StatePending || p.State == StateRunning
}

// IsCompleted returns true if the process has finished (completed, failed, or stopped)
func (p *Process) IsCompleted() bool {
	return p.State == StateCompleted || p.State == StateFailed || p.State == StateStopped
}

// CalculateProgress calculates the progress percentage based on processed vs total files
func (p *Process) CalculateProgress() float64 {
	if p.TotalFiles == 0 {
		return 0.0
	}
	progress := float64(p.ProcessedFiles) / float64(p.TotalFiles) * 100.0
	if progress > 100.0 {
		progress = 100.0
	}
	return progress
}
