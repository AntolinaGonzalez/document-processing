package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"document-processor/internal/models"

	_ "github.com/lib/pq"
)

// DB handles database operations for processes and results.
type DB struct {
	conn *sql.DB
}

// NewDB creates a new DB instance.
func NewDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	return &DB{conn: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) CreateProcess(p *models.Process) error {
	_, err := db.conn.Exec(`INSERT INTO processes (process_id, folder_path, batch_size, state, start_time, total_files, processed_files) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		p.ProcessID, p.FolderPath, p.BatchSize, p.State, p.StartTime, p.TotalFiles, p.ProcessedFiles)
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}
	return nil
}

func (db *DB) UpdateProcessState(processID string, state models.ProcessState, endTime *time.Time, lastError *string) error {
	_, err := db.conn.Exec(`UPDATE processes SET state=$1, end_time=$2, last_error=$3 WHERE process_id=$4`,
		state, endTime, lastError, processID)
	if err != nil {
		return fmt.Errorf("failed to update process state: %w", err)
	}
	return nil
}

func (db *DB) UpdateProcessProgress(processID string, processedFiles, totalFiles int) error {
	_, err := db.conn.Exec(`UPDATE processes SET processed_files=$1, total_files=$2 WHERE process_id=$3`,
		processedFiles, totalFiles, processID)
	if err != nil {
		return fmt.Errorf("failed to update process progress: %w", err)
	}
	return nil
}

func (db *DB) GetProcess(processID string) (*models.Process, error) {
	row := db.conn.QueryRow(`SELECT process_id, folder_path, batch_size, state, start_time, end_time, last_error, total_files, processed_files, created_at, updated_at FROM processes WHERE process_id=$1`, processID)
	var p models.Process
	var endTime sql.NullTime
	var lastError sql.NullString
	var state string
	var createdAt, updatedAt time.Time
	
	if err := row.Scan(&p.ProcessID, &p.FolderPath, &p.BatchSize, &state, &p.StartTime, &endTime, &lastError, &p.TotalFiles, &p.ProcessedFiles, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan process: %w", err)
	}
	
	p.State = models.ProcessState(state)
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	
	if endTime.Valid {
		p.EndTime = &endTime.Time
	}
	if lastError.Valid {
		p.LastError = &lastError.String
	}
	
	// Calculate progress
	p.Progress = p.CalculateProgress()
	
	return &p, nil
}

func (db *DB) ListProcesses() ([]models.Process, error) {
	rows, err := db.conn.Query(`SELECT process_id, folder_path, batch_size, state, start_time, end_time, last_error, total_files, processed_files, created_at, updated_at FROM processes ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query processes: %w", err)
	}
	defer rows.Close()
	
	var processes []models.Process
	for rows.Next() {
		var p models.Process
		var endTime sql.NullTime
		var lastError sql.NullString
		var state string
		var createdAt, updatedAt time.Time
		
		if err := rows.Scan(&p.ProcessID, &p.FolderPath, &p.BatchSize, &state, &p.StartTime, &endTime, &lastError, &p.TotalFiles, &p.ProcessedFiles, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan process row: %w", err)
		}
		
		p.State = models.ProcessState(state)
		p.CreatedAt = createdAt
		p.UpdatedAt = updatedAt
		
		if endTime.Valid {
			p.EndTime = &endTime.Time
		}
		if lastError.Valid {
			p.LastError = &lastError.String
		}
		
		// Calculate progress
		p.Progress = p.CalculateProgress()
		
		processes = append(processes, p)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating process rows: %w", err)
	}
	
	return processes, nil
}

func (db *DB) CreateAnalysisResult(processID string, r *models.FileAnalysisResult) error {
	_, err := db.conn.Exec(`INSERT INTO analysis_results (process_id, file_name, word_count, line_count, char_count, frequent_words, summary) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		processID, r.FileName, r.WordCount, r.LineCount, r.CharCount, toJSON(r.FrequentWords), r.Summary)
	if err != nil {
		return fmt.Errorf("failed to create analysis result: %w", err)
	}
	return nil
}

func (db *DB) GetAnalysisResults(processID string) ([]models.FileAnalysisResult, error) {
	rows, err := db.conn.Query(`SELECT file_name, word_count, line_count, char_count, frequent_words, summary FROM analysis_results WHERE process_id=$1 ORDER BY created_at ASC`, processID)
	if err != nil {
		return nil, fmt.Errorf("failed to query analysis results: %w", err)
	}
	defer rows.Close()
	
	var results []models.FileAnalysisResult
	for rows.Next() {
		var r models.FileAnalysisResult
		var freqWords string
		
		if err := rows.Scan(&r.FileName, &r.WordCount, &r.LineCount, &r.CharCount, &freqWords, &r.Summary); err != nil {
			return nil, fmt.Errorf("failed to scan analysis result row: %w", err)
		}
		
		if err := fromJSON(freqWords, &r.FrequentWords); err != nil {
			return nil, fmt.Errorf("failed to parse frequent words JSON: %w", err)
		}
		
		results = append(results, r)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating analysis result rows: %w", err)
	}
	
	return results, nil
}

// GetResultsSummary returns aggregated results for a process
func (db *DB) GetResultsSummary(processID string) (*models.ResultsSummary, error) {
	rows, err := db.conn.Query(`SELECT file_name, word_count, line_count, frequent_words FROM analysis_results WHERE process_id=$1`, processID)
	if err != nil {
		return nil, fmt.Errorf("failed to query results summary: %w", err)
	}
	defer rows.Close()
	
	var summary models.ResultsSummary
	allFrequentWords := make(map[string]int)
	var fileNames []string
	
	for rows.Next() {
		var fileName string
		var wordCount, lineCount int
		var freqWords string
		
		if err := rows.Scan(&fileName, &wordCount, &lineCount, &freqWords); err != nil {
			return nil, fmt.Errorf("failed to scan result row: %w", err)
		}
		
		summary.TotalWords += wordCount
		summary.TotalLines += lineCount
		fileNames = append(fileNames, fileName)
		
		// Aggregate frequent words
		var fileFreqWords map[string]int
		if err := fromJSON(freqWords, &fileFreqWords); err != nil {
			return nil, fmt.Errorf("failed to parse frequent words: %w", err)
		}
		
		for word, count := range fileFreqWords {
			allFrequentWords[word] += count
		}
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating result rows: %w", err)
	}
	
	// Get top 5 most frequent words across all files
	summary.MostFrequentWords = getTopWords(allFrequentWords, 5)
	summary.FilesProcessed = fileNames
	
	return &summary, nil
}

// Helper functions for JSONB
func toJSON(m map[string]int) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func fromJSON(s string, m *map[string]int) error {
	return json.Unmarshal([]byte(s), m)
}

// getTopWords returns the top N most frequent words as a slice
func getTopWords(freq map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range freq {
		ss = append(ss, kv{k, v})
	}
	
	// Sort by frequency (descending)
	for i := 0; i < len(ss)-1; i++ {
		for j := i + 1; j < len(ss); j++ {
			if ss[i].Value < ss[j].Value {
				ss[i], ss[j] = ss[j], ss[i]
			}
		}
	}
	
	// Return top N words
	var result []string
	for i := 0; i < n && i < len(ss); i++ {
		result = append(result, ss[i].Key)
	}
	return result
}
