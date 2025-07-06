package worker

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"document-processor/internal/db"
	"document-processor/internal/logs"
	"document-processor/internal/models"

	"go.uber.org/zap"
)

// WorkerPool manages a pool of goroutines for processing document batches.
type WorkerPool struct {
	workerCount int
}

// NewWorkerPool creates a new WorkerPool.
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{workerCount: workerCount}
}

// ProcessFolder processes all text files in a folder in batches asynchronously.
func (wp *WorkerPool) ProcessFolder(ctx context.Context, processID, folderPath string, batchSize int, db *db.DB) error {
	logger := logs.GetLogger().With(
		zap.String("processID", processID),
		zap.String("folderPath", folderPath),
		zap.Int("batchSize", batchSize),
	)
	
	logger.Info("Starting folder processing")
	
	files, err := listTextFiles(folderPath)
	if err != nil {
		logger.Error("Failed to list text files", zap.Error(err))
		return err
	}
	
	if len(files) == 0 {
		logger.Warn("No text files found in folder")
		return errors.New("no text files found")
	}
	
	logger.Info("Found text files", zap.Int("fileCount", len(files)))
	
	// Update process with total file count
	if err := db.UpdateProcessProgress(processID, 0, len(files)); err != nil {
		logger.Error("Failed to update process progress", zap.Error(err))
		return err
	}

	// Create batches
	var batches [][]string
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}
		batches = append(batches, files[i:end])
	}

	batchCh := make(chan []string)
	wg := sync.WaitGroup{}
	resultCh := make(chan *models.FileAnalysisResult)
	errCh := make(chan error, 1)
	
	// Progress tracking
	var processedFiles int32

	// Start workers
	for i := 0; i < wp.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerLogger := logger.With(zap.Int("workerID", workerID))
			workerLogger.Debug("Worker started")
			
			for batch := range batchCh {
				for _, file := range batch {
					select {
					case <-ctx.Done():
						workerLogger.Debug("Worker stopped due to context cancellation")
						return
					default:
						// Check if process is paused before processing each file
						if isProcessPaused(processID, db) {
							workerLogger.Debug("Process is paused, waiting...")
							// Wait for process to resume
							for isProcessPaused(processID, db) {
								select {
								case <-ctx.Done():
									workerLogger.Debug("Worker stopped due to context cancellation")
									return
								case <-time.After(1 * time.Second):
									// Check again after 1 second
								}
							}
							workerLogger.Debug("Process resumed, continuing...")
						}
						
						workerLogger.Debug("Processing file", zap.String("file", file))
						res, err := analyzeFile(file)
						if err != nil {
							workerLogger.Error("Failed to analyze file", zap.String("file", file), zap.Error(err))
							errCh <- err
							return
						}
						resultCh <- res
						
						// Update progress
						processed := atomic.AddInt32(&processedFiles, 1)
						if err := db.UpdateProcessProgress(processID, int(processed), len(files)); err != nil {
							workerLogger.Error("Failed to update progress", zap.Error(err))
						}
					}
				}
			}
			workerLogger.Debug("Worker finished")
		}(i)
	}

	// Feed batches
	go func() {
		for _, batch := range batches {
			select {
			case <-ctx.Done():
				logger.Debug("Stopping batch feeding due to context cancellation")
				close(batchCh)
				return
			default:
				// Check if process is paused before feeding batch
				if isProcessPaused(processID, db) {
					logger.Debug("Process is paused, waiting to feed batch...")
					// Wait for process to resume
					for isProcessPaused(processID, db) {
						select {
						case <-ctx.Done():
							logger.Debug("Stopping batch feeding due to context cancellation")
							close(batchCh)
							return
						case <-time.After(1 * time.Second):
							// Check again after 1 second
						}
					}
					logger.Debug("Process resumed, feeding batch...")
				}
				batchCh <- batch
			}
		}
		close(batchCh)
		logger.Debug("Finished feeding batches")
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Collect results and store in DB
	processed := 0
	for {
		select {
		case <-ctx.Done():
			logger.Info("Process stopped by user")
			return errors.New("process stopped")
		case err := <-errCh:
			logger.Error("Processing error occurred", zap.Error(err))
			return err
		case res := <-resultCh:
			if res != nil {
				if err := db.CreateAnalysisResult(processID, res); err != nil {
					logger.Error("Failed to store analysis result", zap.Error(err))
					return err
				}
				processed++
				logger.Debug("Stored analysis result", zap.String("fileName", res.FileName))
			}
		case <-done:
			logger.Info("Processing completed", zap.Int("processedFiles", processed))
			return nil
		}
	}
}

// isProcessPaused checks if a process is currently paused
func isProcessPaused(processID string, db *db.DB) bool {
	proc, err := db.GetProcess(processID)
	if err != nil {
		return false
	}
	return proc.State == models.StatePaused
}

// listTextFiles returns a list of .txt files in the folder (recursively).
func listTextFiles(folder string) ([]string, error) {
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".txt") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// analyzeFile reads a file and returns its analysis result.
func analyzeFile(path string) (*models.FileAnalysisResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	wordCount := 0
	charCount := 0
	wordFreq := make(map[string]int)
	var content strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		words := strings.Fields(line)
		wordCount += len(words)
		charCount += len(line)
		for _, w := range words {
			w = strings.ToLower(strings.Trim(w, ",.;:!?()[]{}\"'`"))
			if w != "" {
				wordFreq[w]++
			}
		}
		content.WriteString(line + " ")
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	summary := generateSummary(content.String())

	return &models.FileAnalysisResult{
		FileName:      filepath.Base(path),
		WordCount:     wordCount,
		LineCount:     lineCount,
		CharCount:     charCount,
		FrequentWords: topNWords(wordFreq, 5),
		Summary:       summary,
	}, nil
}

// topNWords returns the N most frequent words as a map.
func topNWords(freq map[string]int, n int) map[string]int {
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range freq {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	result := make(map[string]int)
	for i := 0; i < n && i < len(ss); i++ {
		result[ss[i].Key] = ss[i].Value
	}
	return result
}

// generateSummary returns the first 2 sentences as a summary.
func generateSummary(text string) string {
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	if len(sentences) == 0 {
		return ""
	}
	if len(sentences) > 2 {
		return strings.TrimSpace(sentences[0]) + ". " + strings.TrimSpace(sentences[1]) + "."
	}
	return strings.TrimSpace(sentences[0]) + "."
}
