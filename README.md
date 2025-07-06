# Document Processing System (GoLang)

## Overview
A robust, asynchronous document processing system with a REST API for loading, processing, and analyzing text files. Built with Go, Gin, PostgreSQL, and idiomatic concurrency patterns.

## Features
- **Asynchronous Processing**: Concurrent batch processing of text files using GoRoutines
- **REST API**: Complete HTTP API to start, stop, monitor, and retrieve results
- **Text Analysis**: Extracts word count, line count, character count, frequent words, and summaries
- **Progress Tracking**: Real-time progress monitoring with percentage completion
- **Persistent State**: PostgreSQL database for process metadata and analysis results
- **Structured Logging**: Comprehensive logging with Zap for observability
- **Graceful Shutdown**: Proper cleanup and resource management
- **Error Handling**: Robust error handling with detailed error messages

## Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │ Process Manager │    │  Worker Pool    │
│   (Gin)         │◄──►│   (Orchestrator)│◄──►│  (GoRoutines)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Models        │    │   Database      │    │   File System   │
│   (Structs)     │    │  (PostgreSQL)   │    │   (Text Files)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Components
- **API Layer** (`internal/api/`): HTTP request handling and response formatting
- **Process Manager** (`internal/manager/`): Process lifecycle orchestration with state management
- **Worker Pool** (`internal/worker/`): Concurrent file processing with progress tracking
- **Database Layer** (`internal/db/`): Data persistence and retrieval with connection pooling
- **Models** (`internal/models/`): Data structures and state definitions
- **Logging** (`internal/logs/`): Structured logging with Zap

## Project Structure
```
cmd/server/         # Main server entrypoint with graceful shutdown
internal/api/       # REST API handlers with structured logging
internal/manager/   # Process manager logic with state transitions
internal/worker/    # Worker pool and processing logic with progress tracking
internal/db/        # Database access layer with connection pooling
internal/models/    # Data structures and state definitions
internal/logs/      # Structured logging setup with Zap
configs/            # Configuration files (YAML)
migrations/         # SQL migration scripts
```

## Setup

### Prerequisites
1. **Docker** and **Docker Compose** (recommended)
   - [Docker Desktop](https://www.docker.com/products/docker-desktop/) for macOS/Windows
   - [Docker Engine](https://docs.docker.com/engine/install/) for Linux
2. **Git** for cloning the repository

### Quick Start with Docker (Recommended)

1. **Clone this repository:**
   ```bash
   git clone <repository-url>
   cd project
   ```

2. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

3. **Start the application with Docker Compose:**
   ```bash
   docker-compose up -d
   ```

4. **Verify the service is running:**
   ```bash
   curl http://localhost:8080/health
   ```

5. **Run the demo to test the system:**
   ```bash
   ./demo-test.sh
   ```

### Manual Setup (Alternative)

If you prefer to run without Docker:

1. **Prerequisites:**
   - Go 1.23.0+
   - PostgreSQL 12+

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Set up PostgreSQL database:**
   ```bash
   createdb document_processor
   ```

4. **Run database migrations:**
   ```bash
   psql -d document_processor -f migrations/001_initial_schema.sql
   psql -d document_processor -f migrations/002_add_progress_columns.sql
   ```

5. **Configure the application in `configs/config.yaml`**

6. **Start the server:**
   ```bash
   go run cmd/server/main.go
   ```

## Configuration

### Docker Setup (Default)
The application is pre-configured for Docker deployment. The `configs/config.yaml` file is set up for containerized execution:

```yaml
server:
  port: 8080

database:
  host: postgres  # Docker service name
  port: 5432
  user: postgres
  password: password
  dbname: document_processor
  sslmode: disable

logging:
  level: info  # debug, info, warn, error
```

### Local Setup
If running without Docker, update the database host in `configs/config.yaml`:

```yaml
database:
  host: localhost  # Change from 'postgres' to 'localhost'
  port: 5432
  user: postgres
  password: password
  dbname: document_processor
  sslmode: disable
```

## API Endpoints

### Start Processing
```http
POST /process/start
Content-Type: application/json

{
  "folderPath": "/path/to/text/files",
  "batchSize": 10
}
```

**Response:**
```json
{
  "process_id": "uuid-string",
  "status": "PENDING"
}
```

### Stop Processing
```http
POST /process/stop/{process_id}
```

**Response:**
```json
{
  "process_id": "uuid-string",
  "status": "STOPPED",
  "message": "Process stopped"
}
```

### Get Process Status
```http
GET /process/status/{process_id}
```

**Response:**
```json
{
  "process_id": "uuid-string",
  "state": "RUNNING",
  "progress": 45.5,
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": null,
  "error_detail": null
}
```

### List All Processes
```http
GET /process/list
```

**Response:**
```json
{
  "processes": [
    {
      "process_id": "uuid-string",
      "state": "COMPLETED",
      "progress": 100.0,
      "start_time": "2024-01-01T10:00:00Z",
      "end_time": "2024-01-01T10:05:00Z"
    }
  ]
}
```

### Get Analysis Results
```http
GET /process/results/{process_id}
```

**Response:**
```json
{
  "process_id": "uuid-string",
  "results": [
    {
      "file_name": "document.txt",
      "word_count": 1500,
      "line_count": 75,
      "char_count": 8500,
      "frequent_words": {
        "the": 45,
        "and": 32,
        "to": 28
      },
      "summary": "This document discusses..."
    }
  ]
}
```

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1704110400
}
```

## Process States
- **PENDING**: Process created, waiting to start
- **RUNNING**: Currently processing files
- **PAUSED**: Temporarily paused (not implemented)
- **COMPLETED**: Successfully finished processing
- **FAILED**: Encountered an error during processing
- **STOPPED**: Manually stopped by user

## Error Handling
All endpoints return consistent error responses:
```json
{
  "error": "Error description",
  "details": "Additional error details"
}
```

## Logging
The system uses structured logging with Zap. Log levels can be configured:
- **debug**: Detailed debugging information
- **info**: General operational information
- **warn**: Warning messages
- **error**: Error conditions

## Performance
- **Concurrent Processing**: Configurable worker pool for parallel file processing
- **Batch Processing**: Configurable batch sizes for optimal performance
- **Connection Pooling**: Database connection pooling for efficient resource usage
- **Memory Efficient**: Streaming file processing with buffered I/O

## Development

### Testing and Demo

#### Run the Demo Script
```bash
# Make sure the service is running first
docker-compose up -d

# Run the comprehensive demo
./demo-test.sh
```

#### Manual Testing
```bash
# Health check
curl http://localhost:8080/health

# Start processing
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{"folder_path": "/data", "batch_size": 2}'

# Check status (replace with actual process_id)
curl http://localhost:8080/process/status/YOUR_PROCESS_ID

# Get results
curl http://localhost:8080/process/results/YOUR_PROCESS_ID
```

#### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o bin/server cmd/server/main.go
```

### Docker Support

#### Using Docker Compose (Recommended)
```bash
# Start all services (app + database)
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop all services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

#### Using Docker directly
```bash
# Build the image
docker build -t document-processor .

# Run with PostgreSQL (requires external database)
docker run -p 8080:8080 \
  -e DB_HOST=your-postgres-host \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=password \
  -e DB_NAME=document_processor \
  document-processor
```

#### Docker Commands Reference
```bash
# Check container status
docker-compose ps

# View application logs
docker-compose logs app

# View database logs
docker-compose logs postgres

# Execute commands in running container
docker-compose exec app sh

# Access database directly
docker-compose exec postgres psql -U postgres -d document_processor
```

## Troubleshooting

### Common Issues

#### Docker Issues
```bash
# If containers fail to start, check logs
docker-compose logs

# If database connection fails, restart services
docker-compose down
docker-compose up -d

# If you need to reset the database
docker-compose down -v  # This removes volumes
docker-compose up -d
```

#### Database Migration Issues
If you encounter database schema errors:
```bash
# Apply missing migrations manually
docker-compose exec postgres psql -U postgres -d document_processor -f /docker-entrypoint-initdb.d/001_initial_schema.sql
docker-compose exec postgres psql -U postgres -d document_processor -f /docker-entrypoint-initdb.d/002_add_progress_columns.sql
```

#### Port Conflicts
If port 8080 is already in use:
```bash
# Check what's using the port
lsof -i :8080

# Or modify docker-compose.yml to use a different port
# Change "8080:8080" to "8081:8080"
```

### Getting Help
- Check the logs: `docker-compose logs -f app`
- Verify service health: `docker-compose ps`
- Test the API: `curl http://localhost:8080/health`

## Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License
MIT License - see LICENSE file for details.
