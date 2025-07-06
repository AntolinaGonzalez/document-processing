package manager

import (
	"context"
	"errors"
	"sync"
	"time"

	"document-processor/internal/db"
	"document-processor/internal/logs"
	"document-processor/internal/models"
	"document-processor/internal/worker"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProcessManager manages the lifecycle of document processing jobs.
type ProcessManager struct {
	db         *db.DB
	workerPool *worker.WorkerPool
	mu         sync.Mutex
	processCtx map[string]context.CancelFunc // processID -> cancel func
	pausedCtx  map[string]context.Context   // processID -> paused context
}

// NewProcessManager creates a new ProcessManager.
func NewProcessManager(db *db.DB, workerPool *worker.WorkerPool) *ProcessManager {
	return &ProcessManager{
		db:         db,
		workerPool: workerPool,
		processCtx: make(map[string]context.CancelFunc),
		pausedCtx:  make(map[string]context.Context),
	}
}

// StartProcess starts a new document processing job.
func (pm *ProcessManager) StartProcess(folderPath string, batchSize int) (string, error) {
	logger := logs.GetLogger().With(
		zap.String("folder_path", folderPath),
		zap.Int("batch_size", batchSize),
	)
	
	logger.Info("Starting new document processing job")
	
	processID := uuid.New().String()
	proc := &models.Process{
		ProcessID:      processID,
		FolderPath:     folderPath,
		BatchSize:      batchSize,
		State:          models.StatePending,
		StartTime:      time.Now(),
		TotalFiles:     0,
		ProcessedFiles: 0,
	}
	
	if err := pm.db.CreateProcess(proc); err != nil {
		logger.Error("Failed to create process in database", zap.Error(err))
		return "", err
	}

	ctx, cancel := context.WithCancel(context.Background())
	pm.mu.Lock()
	pm.processCtx[processID] = cancel
	pm.mu.Unlock()

	logger.Info("Process created successfully", zap.String("process_id", processID))
	
	go pm.runProcess(ctx, proc)
	return processID, nil
}

func (pm *ProcessManager) runProcess(ctx context.Context, proc *models.Process) {
	logger := logs.GetLogger().With(zap.String("process_id", proc.ProcessID))
	
	// Update state to RUNNING
	if err := pm.db.UpdateProcessState(proc.ProcessID, models.StateRunning, nil, nil); err != nil {
		logger.Error("Failed to update process state to RUNNING", zap.Error(err))
		return
	}
	
	logger.Info("Process started running")

	err := pm.workerPool.ProcessFolder(ctx, proc.ProcessID, proc.FolderPath, proc.BatchSize, pm.db)
	endTime := time.Now()
	
	if err != nil {
		errStr := err.Error()
		logger.Error("Process failed", zap.Error(err))
		if updateErr := pm.db.UpdateProcessState(proc.ProcessID, models.StateFailed, &endTime, &errStr); updateErr != nil {
			logger.Error("Failed to update process state to FAILED", zap.Error(updateErr))
		}
	} else {
		logger.Info("Process completed successfully")
		if updateErr := pm.db.UpdateProcessState(proc.ProcessID, models.StateCompleted, &endTime, nil); updateErr != nil {
			logger.Error("Failed to update process state to COMPLETED", zap.Error(updateErr))
		}
	}

	pm.mu.Lock()
	delete(pm.processCtx, proc.ProcessID)
	pm.mu.Unlock()
	
	logger.Info("Process cleanup completed")
}

// PauseProcess pauses a running process.
func (pm *ProcessManager) PauseProcess(processID string) error {
	logger := logs.GetLogger().With(zap.String("process_id", processID))
	
	logger.Info("Attempting to pause process")
	
	// Check if process exists and is running
	proc, err := pm.db.GetProcess(processID)
	if err != nil {
		logger.Error("Process not found", zap.Error(err))
		return errors.New("process not found")
	}
	
	if proc.State != models.StateRunning {
		logger.Warn("Process is not running, cannot pause", zap.String("state", string(proc.State)))
		return errors.New("process is not running")
	}
	
	// Create a paused context that can be cancelled to resume
	pausedCtx, _ := context.WithCancel(context.Background())
	
	pm.mu.Lock()
	pm.pausedCtx[processID] = pausedCtx
	pm.mu.Unlock()
	
	// Update database state
	if err := pm.db.UpdateProcessState(processID, models.StatePaused, nil, nil); err != nil {
		logger.Error("Failed to update process state to PAUSED", zap.Error(err))
		return err
	}
	
	logger.Info("Process paused successfully")
	return nil
}

// ResumeProcess resumes a paused process.
func (pm *ProcessManager) ResumeProcess(processID string) error {
	logger := logs.GetLogger().With(zap.String("process_id", processID))
	
	logger.Info("Attempting to resume process")
	
	// Check if process exists and is paused
	proc, err := pm.db.GetProcess(processID)
	if err != nil {
		logger.Error("Process not found", zap.Error(err))
		return errors.New("process not found")
	}
	
	if proc.State != models.StatePaused {
		logger.Warn("Process is not paused, cannot resume", zap.String("state", string(proc.State)))
		return errors.New("process is not paused")
	}
	
	// Cancel the paused context to resume processing
	pm.mu.Lock()
	pausedCtx, exists := pm.pausedCtx[processID]
	if exists {
		// Cancel the paused context to signal resume
		select {
		case <-pausedCtx.Done():
			// Context already cancelled, create a new one
		default:
			// Cancel the context to resume
		}
		delete(pm.pausedCtx, processID)
	}
	pm.mu.Unlock()
	
	// Update database state
	if err := pm.db.UpdateProcessState(processID, models.StateRunning, nil, nil); err != nil {
		logger.Error("Failed to update process state to RUNNING", zap.Error(err))
		return err
	}
	
	logger.Info("Process resumed successfully")
	return nil
}

// StopProcess stops a running process.
func (pm *ProcessManager) StopProcess(processID string) error {
	logger := logs.GetLogger().With(zap.String("process_id", processID))
	
	logger.Info("Attempting to stop process")
	
	pm.mu.Lock()
	cancel, ok := pm.processCtx[processID]
	pm.mu.Unlock()
	
	if !ok {
		logger.Warn("Process not found in active processes")
		return errors.New("process not running or already stopped")
	}
	
	cancel()
	endTime := time.Now()
	
	if err := pm.db.UpdateProcessState(processID, models.StateStopped, &endTime, nil); err != nil {
		logger.Error("Failed to update process state to STOPPED", zap.Error(err))
		return err
	}
	
	logger.Info("Process stopped successfully")
	return nil
}

// GetProcessStatus returns the status of a process.
func (pm *ProcessManager) GetProcessStatus(processID string) (*models.Process, error) {
	logger := logs.GetLogger().With(zap.String("process_id", processID))
	
	proc, err := pm.db.GetProcess(processID)
	if err != nil {
		logger.Error("Failed to get process status", zap.Error(err))
		return nil, err
	}
	
	logger.Debug("Retrieved process status", 
		zap.String("state", string(proc.State)),
		zap.Float64("progress", proc.Progress))
	
	return proc, nil
}

// ListProcesses lists all processes.
func (pm *ProcessManager) ListProcesses() ([]models.Process, error) {
	logger := logs.GetLogger()
	
	procs, err := pm.db.ListProcesses()
	if err != nil {
		logger.Error("Failed to list processes", zap.Error(err))
		return nil, err
	}
	
	logger.Debug("Retrieved process list", zap.Int("count", len(procs)))
	return procs, nil
}

// GetProcessResults returns analysis results for a process.
func (pm *ProcessManager) GetProcessResults(processID string) ([]models.FileAnalysisResult, error) {
	logger := logs.GetLogger().With(zap.String("process_id", processID))
	
	results, err := pm.db.GetAnalysisResults(processID)
	if err != nil {
		logger.Error("Failed to get process results", zap.Error(err))
		return nil, err
	}
	
	logger.Debug("Retrieved process results", zap.Int("result_count", len(results)))
	return results, nil
}

// GetProcessResultsSummary returns aggregated results summary for a process.
func (pm *ProcessManager) GetProcessResultsSummary(processID string) (*models.ResultsSummary, error) {
	logger := logs.GetLogger().With(zap.String("process_id", processID))
	
	summary, err := pm.db.GetResultsSummary(processID)
	if err != nil {
		logger.Error("Failed to get results summary", zap.Error(err))
		return nil, err
	}
	
	logger.Debug("Retrieved results summary", 
		zap.Int("total_words", summary.TotalWords),
		zap.Int("total_lines", summary.TotalLines),
		zap.Int("files_processed", len(summary.FilesProcessed)))
	
	return summary, nil
}
