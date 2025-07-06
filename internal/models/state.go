package models

import "time"

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type ProgressInfo struct {
	TotalFiles     int `json:"total_files"`
	ProcessedFiles int `json:"processed_files"`
	Percentage     int `json:"percentage"`
}

type ResultsSummary struct {
	TotalWords           int      `json:"total_words"`
	TotalLines           int      `json:"total_lines"`
	MostFrequentWords    []string `json:"most_frequent_words"`
	FilesProcessed       []string `json:"files_processed"`
}

type ProcessStatusResponse struct {
	ProcessID            string         `json:"process_id"`
	Status               string         `json:"status"`
	Progress             ProgressInfo   `json:"progress"`
	StartedAt            time.Time      `json:"started_at"`
	EstimatedCompletion  *time.Time     `json:"estimated_completion,omitempty"`
	Results              *ResultsSummary `json:"results,omitempty"`
	ErrorDetail          *string        `json:"error_detail,omitempty"`
}

type ProcessSummary struct {
	ProcessID           string         `json:"process_id"`
	Status              string         `json:"status"`
	Progress            ProgressInfo   `json:"progress"`
	StartedAt           time.Time      `json:"started_at"`
	EstimatedCompletion *time.Time     `json:"estimated_completion,omitempty"`
}
