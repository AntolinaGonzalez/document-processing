package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"io/ioutil"

	"gopkg.in/yaml.v2"

	"document-processor/internal/api"
	"document-processor/internal/db"
	"document-processor/internal/logs"
	"document-processor/internal/manager"
	"document-processor/internal/worker"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &cfg, nil
}

func main() {
	// Load configuration
	cfg, err := loadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured logging
	if err := logs.InitLogger(cfg.Logging.Level); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logs.Sync()

	logger := logs.GetLogger()
	logger.Info("Starting Document Processing System")

	// Build database connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	// Connect to database
	dbConn, err := db.NewDB(dsn)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbConn.Close()

	logger.Info("Connected to database successfully")

	// Initialize components
	workerPool := worker.NewWorkerPool(4) // You can make this configurable
	pm := manager.NewProcessManager(dbConn, workerPool)
	handler := &api.APIHandler{PM: pm}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode) // Set to gin.DebugMode for development
	r := gin.New()
	
	// Add middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// API endpoints
	r.POST("/process/start", handler.StartProcessHandler)
	r.POST("/process/stop/:process_id", handler.StopProcessHandler)
	r.POST("/process/pause/:process_id", handler.PauseProcessHandler)
	r.POST("/process/resume/:process_id", handler.ResumeProcessHandler)
	r.GET("/process/status/:process_id", handler.GetProcessStatusHandler)
	r.GET("/process/list", handler.ListProcessesHandler)
	r.GET("/process/results/:process_id", handler.GetProcessResultsHandler)

	// Determine port
	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
