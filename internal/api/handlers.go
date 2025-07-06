package api

import (
	"net/http"
	"time"

	"document-processor/internal/logs"
	"document-processor/internal/manager"
	"document-processor/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// APIHandler holds the process manager reference.
type APIHandler struct {
	PM *manager.ProcessManager
}

// POST /process/start
// Request: { folder_path: string, batch_size: int }
type StartProcessRequest struct {
	FolderPath string `json:"folder_path" binding:"required"`
	BatchSize  int    `json:"batch_size" binding:"required,min=1"`
}
type StartProcessResponse struct {
	ProcessID string `json:"process_id"`
	Status    string `json:"status"`
}

func (h *APIHandler) StartProcessHandler(c *gin.Context) {
	logger := logs.GetLogger()
	
	var req StartProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid start process request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request", Details: err.Error()})
		return
	}
	
	logger.Info("Received start process request", 
		zap.String("folder_path", req.FolderPath),
		zap.Int("batch_size", req.BatchSize))
	
	pid, err := h.PM.StartProcess(req.FolderPath, req.BatchSize)
	if err != nil {
		logger.Error("Failed to start process", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to start process", Details: err.Error()})
		return
	}
	
	logger.Info("Process started successfully", zap.String("process_id", pid))
	c.JSON(http.StatusOK, StartProcessResponse{ProcessID: pid, Status: string(models.StatePending)})
}

// POST /process/stop/:process_id
type StopProcessResponse struct {
	ProcessID string `json:"process_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func (h *APIHandler) StopProcessHandler(c *gin.Context) {
	pid := c.Param("process_id")
	if pid == "" {
		logger := logs.GetLogger()
		logger.Error("Missing process_id in stop request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing process_id"})
		return
	}
	
	logger := logs.GetLogger().With(zap.String("process_id", pid))
	logger.Info("Received stop process request")
	
	err := h.PM.StopProcess(pid)
	if err != nil {
		logger.Error("Failed to stop process", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "failed to stop process", Details: err.Error()})
		return
	}
	
	logger.Info("Process stopped successfully")
	c.JSON(http.StatusOK, StopProcessResponse{ProcessID: pid, Status: string(models.StateStopped), Message: "Process stopped"})
}

// POST /process/pause/:process_id
type PauseProcessResponse struct {
	ProcessID string `json:"process_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func (h *APIHandler) PauseProcessHandler(c *gin.Context) {
	pid := c.Param("process_id")
	if pid == "" {
		logger := logs.GetLogger()
		logger.Error("Missing process_id in pause request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing process_id"})
		return
	}
	
	logger := logs.GetLogger().With(zap.String("process_id", pid))
	logger.Info("Received pause process request")
	
	err := h.PM.PauseProcess(pid)
	if err != nil {
		logger.Error("Failed to pause process", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "failed to pause process", Details: err.Error()})
		return
	}
	
	logger.Info("Process paused successfully")
	c.JSON(http.StatusOK, PauseProcessResponse{ProcessID: pid, Status: string(models.StatePaused), Message: "Process paused"})
}

// POST /process/resume/:process_id
type ResumeProcessResponse struct {
	ProcessID string `json:"process_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func (h *APIHandler) ResumeProcessHandler(c *gin.Context) {
	pid := c.Param("process_id")
	if pid == "" {
		logger := logs.GetLogger()
		logger.Error("Missing process_id in resume request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing process_id"})
		return
	}
	
	logger := logs.GetLogger().With(zap.String("process_id", pid))
	logger.Info("Received resume process request")
	
	err := h.PM.ResumeProcess(pid)
	if err != nil {
		logger.Error("Failed to resume process", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "failed to resume process", Details: err.Error()})
		return
	}
	
	logger.Info("Process resumed successfully")
	c.JSON(http.StatusOK, ResumeProcessResponse{ProcessID: pid, Status: string(models.StateRunning), Message: "Process resumed"})
}

// GET /process/status/:process_id
func (h *APIHandler) GetProcessStatusHandler(c *gin.Context) {
	pid := c.Param("process_id")
	if pid == "" {
		logger := logs.GetLogger()
		logger.Error("Missing process_id in status request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing process_id"})
		return
	}
	
	logger := logs.GetLogger().With(zap.String("process_id", pid))
	logger.Debug("Received get process status request")
	
	proc, err := h.PM.GetProcessStatus(pid)
	if err != nil {
		logger.Error("Failed to get process status", zap.Error(err))
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "process not found", Details: err.Error()})
		return
	}
	
	// Calculate progress info
	progress := models.ProgressInfo{
		TotalFiles:     proc.TotalFiles,
		ProcessedFiles: proc.ProcessedFiles,
		Percentage:     int(proc.Progress),
	}
	
	// Calculate estimated completion time
	var estimatedCompletion *time.Time
	if proc.State == models.StateRunning && proc.ProcessedFiles > 0 && proc.TotalFiles > 0 {
		elapsed := time.Since(proc.StartTime)
		avgTimePerFile := elapsed / time.Duration(proc.ProcessedFiles)
		remainingFiles := proc.TotalFiles - proc.ProcessedFiles
		estimatedTime := time.Now().Add(avgTimePerFile * time.Duration(remainingFiles))
		estimatedCompletion = &estimatedTime
	}
	
	// Build response
	resp := models.ProcessStatusResponse{
		ProcessID:           proc.ProcessID,
		Status:              string(proc.State),
		Progress:            progress,
		StartedAt:           proc.StartTime,
		EstimatedCompletion: estimatedCompletion,
		ErrorDetail:         proc.LastError,
	}
	
	// Add results if process is completed
	if proc.State == models.StateCompleted {
		resultsSummary, err := h.PM.GetProcessResultsSummary(pid)
		if err != nil {
			logger.Error("Failed to get results summary", zap.Error(err))
		} else {
			resp.Results = resultsSummary
		}
	}
	
	logger.Debug("Returning process status", 
		zap.String("status", resp.Status),
		zap.Int("percentage", resp.Progress.Percentage))
	
	c.JSON(http.StatusOK, resp)
}

// GET /process/list
func (h *APIHandler) ListProcessesHandler(c *gin.Context) {
	logger := logs.GetLogger()
	logger.Debug("Received list processes request")
	
	procs, err := h.PM.ListProcesses()
	if err != nil {
		logger.Error("Failed to list processes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list processes", Details: err.Error()})
		return
	}
	
	var summaries []models.ProcessSummary
	for _, p := range procs {
		// Calculate progress info
		progress := models.ProgressInfo{
			TotalFiles:     p.TotalFiles,
			ProcessedFiles: p.ProcessedFiles,
			Percentage:     int(p.Progress),
		}
		
		// Calculate estimated completion time
		var estimatedCompletion *time.Time
		if p.State == models.StateRunning && p.ProcessedFiles > 0 && p.TotalFiles > 0 {
			elapsed := time.Since(p.StartTime)
			avgTimePerFile := elapsed / time.Duration(p.ProcessedFiles)
			remainingFiles := p.TotalFiles - p.ProcessedFiles
			estimatedTime := time.Now().Add(avgTimePerFile * time.Duration(remainingFiles))
			estimatedCompletion = &estimatedTime
		}
		
		summaries = append(summaries, models.ProcessSummary{
			ProcessID:           p.ProcessID,
			Status:              string(p.State),
			Progress:            progress,
			StartedAt:           p.StartTime,
			EstimatedCompletion: estimatedCompletion,
		})
	}
	
	logger.Debug("Returning process list", zap.Int("count", len(summaries)))
	c.JSON(http.StatusOK, gin.H{"processes": summaries})
}

// GET /process/results/:process_id
func (h *APIHandler) GetProcessResultsHandler(c *gin.Context) {
	pid := c.Param("process_id")
	if pid == "" {
		logger := logs.GetLogger()
		logger.Error("Missing process_id in results request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing process_id"})
		return
	}
	
	logger := logs.GetLogger().With(zap.String("process_id", pid))
	logger.Debug("Received get process results request")
	
	results, err := h.PM.GetProcessResults(pid)
	if err != nil {
		logger.Error("Failed to get process results", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to get results", Details: err.Error()})
		return
	}
	
	logger.Debug("Returning process results", zap.Int("result_count", len(results)))
	c.JSON(http.StatusOK, gin.H{"process_id": pid, "results": results})
}
