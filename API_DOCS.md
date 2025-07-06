# Document Processing System API Documentation

## Table of Contents
1. [Overview](#overview)
2. [Base URL and Authentication](#base-url-and-authentication)
3. [Error Handling](#error-handling)
4. [Data Types](#data-types)
5. [Endpoints](#endpoints)
   - [Health Check](#health-check)
   - [Start Document Processing](#start-document-processing)
   - [Stop Document Processing](#stop-document-processing)
   - [Get Process Status](#get-process-status)
   - [List All Processes](#list-all-processes)
   - [Get Analysis Results](#get-analysis-results)
6. [Process States](#process-states)
7. [Usage Examples](#usage-examples)
8. [Performance Considerations](#performance-considerations)

## Overview

The Document Processing System provides a REST API for asynchronously processing text files and extracting comprehensive analysis results. The system processes files in batches using concurrent workers and provides real-time progress tracking.

**Key Features:**
- Asynchronous batch processing with configurable worker pools
- Real-time progress tracking with estimated completion times
- Comprehensive text analysis (word count, line count, character count, frequent words)
- Persistent state management with PostgreSQL
- Structured error responses and logging

## Base URL and Authentication

**Base URL:** `http://localhost:8080`

**Authentication:** Currently, the API does not require authentication. In production environments, consider implementing API keys or JWT tokens.

## Error Handling

All endpoints return consistent error responses in the following format:

```json
{
  "error": "Error description",
  "details": "Additional error details"
}
```

**Common HTTP Status Codes:**
- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters or validation errors
- `404 Not Found` - Resource not found (e.g., process ID doesn't exist)
- `500 Internal Server Error` - Server error or processing failure

## Data Types

### Process State
```typescript
type ProcessState = "PENDING" | "RUNNING" | "PAUSED" | "COMPLETED" | "FAILED" | "STOPPED"
```

### Progress Information
```typescript
interface ProgressInfo {
  total_files: number;
  processed_files: number;
  percentage: number;
}
```

### Results Summary
```typescript
interface ResultsSummary {
  total_words: number;
  total_lines: number;
  most_frequent_words: string[];
  files_processed: string[];
}
```



## Endpoints

### Health Check

Check if the service is running and healthy.

**Endpoint:** `GET /health`

**Description:** Returns the current health status of the service.

**Request:** No request body required.

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1704110400
}
```

**Example:**
```bash
curl -X GET http://localhost:8080/health
```

**Response Example:**
```json
{
  "status": "ok",
  "timestamp": 1751822489
}
```

---

### Start Document Processing

Start a new document processing job for files in a specified folder.

**Endpoint:** `POST /process/start`

**Description:** Initiates asynchronous processing of all `.txt` files in the specified folder.

**Request Body:**
```json
{
  "folder_path": "string",
  "batch_size": "integer"
}
```

**Parameters:**
| Parameter | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `folder_path` | string | Yes | Path to folder containing `.txt` files | Must be a valid directory path |
| `batch_size` | integer | Yes | Number of files to process in each batch | Must be >= 1 |

**Response:**
```json
{
  "process_id": "string",
  "status": "string"
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `process_id` | string | Unique identifier for the processing job (UUID) |
| `status` | string | Initial status of the process (always "PENDING") |

**Example:**
```bash
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{
    "folder_path": "/data",
    "batch_size": 5
  }'
```

**Success Response (200):**
```json
{
  "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
  "status": "PENDING"
}
```

**Error Response (400):**
```json
{
  "error": "invalid request",
  "details": "Key: 'StartProcessRequest.FolderPath' Error:Field validation for 'FolderPath' failed on the 'required' tag"
}
```

---

### Stop Document Processing

Stop a running document processing job.

**Endpoint:** `POST /process/stop/{process_id}`

**Description:** Gracefully stops a running processing job and updates its state to STOPPED.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `process_id` | string | Yes | UUID of the process to stop |

**Request:** No request body required.

**Response:**
```json
{
  "process_id": "string",
  "status": "string",
  "message": "string"
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `process_id` | string | The process ID that was stopped |
| `status` | string | New status of the process (always "STOPPED") |
| `message` | string | Confirmation message |

**Example:**
```bash
curl -X POST http://localhost:8080/process/stop/5ffce937-a5ce-4e37-a65f-4d42ca28a69b
```

**Success Response (200):**
```json
{
  "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
  "status": "STOPPED",
  "message": "Process stopped"
}
```

**Error Response (400):**
```json
{
  "error": "failed to stop process",
  "details": "process not running or already stopped"
}
```

---

### Pause Document Processing

Pause a running document processing job.

**Endpoint:** `POST /process/pause/{process_id}`

**Description:** Temporarily pauses a running processing job. The process can be resumed later.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `process_id` | string | Yes | UUID of the process to pause |

**Request:** No request body required.

**Response:**
```json
{
  "process_id": "string",
  "status": "string",
  "message": "string"
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `process_id` | string | The process ID that was paused |
| `status` | string | New status of the process (always "PAUSED") |
| `message` | string | Confirmation message |

**Example:**
```bash
curl -X POST http://localhost:8080/process/pause/5ffce937-a5ce-4e37-a65f-4d42ca28a69b
```

**Success Response (200):**
```json
{
  "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
  "status": "PAUSED",
  "message": "Process paused"
}
```

**Error Response (400):**
```json
{
  "error": "failed to pause process",
  "details": "process is not running"
}
```

---

### Resume Document Processing

Resume a paused document processing job.

**Endpoint:** `POST /process/resume/{process_id}`

**Description:** Resumes a paused processing job from where it left off.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `process_id` | string | Yes | UUID of the process to resume |

**Request:** No request body required.

**Response:**
```json
{
  "process_id": "string",
  "status": "string",
  "message": "string"
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `process_id` | string | The process ID that was resumed |
| `status` | string | New status of the process (always "RUNNING") |
| `message` | string | Confirmation message |

**Example:**
```bash
curl -X POST http://localhost:8080/process/resume/5ffce937-a5ce-4e37-a65f-4d42ca28a69b
```

**Success Response (200):**
```json
{
  "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
  "status": "RUNNING",
  "message": "Process resumed"
}
```

**Error Response (400):**
```json
{
  "error": "failed to resume process",
  "details": "process is not paused"
}
```

---

### Get Process Status

Get the current status and progress of a document processing job.

**Endpoint:** `GET /process/status/{process_id}`

**Description:** Returns detailed information about a processing job including current state, progress, and estimated completion time.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `process_id` | string | Yes | UUID of the process |

**Request:** No request body required.

**Response:**
```json
{
  "process_id": "string",
  "status": "string",
  "progress": {
    "total_files": "integer",
    "processed_files": "integer",
    "percentage": "integer"
  },
  "started_at": "string",
  "estimated_completion": "string",
  "results": {
    "total_words": "integer",
    "total_lines": "integer",
    "most_frequent_words": ["string"],
    "files_processed": ["string"]
  },
  "error_detail": "string"
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `process_id` | string | Unique identifier for the processing job |
| `status` | string | Current state of the process |
| `progress` | object | Progress information |
| `progress.total_files` | integer | Total number of files to process |
| `progress.processed_files` | integer | Number of files processed so far |
| `progress.percentage` | integer | Completion percentage (0-100) |
| `started_at` | string | ISO 8601 timestamp when processing started |
| `estimated_completion` | string | ISO 8601 timestamp of estimated completion (only for RUNNING state) |
| `results` | object | Aggregated results summary (only for COMPLETED state) |
| `error_detail` | string | Error message if process failed (only for FAILED state) |

**Example:**
```bash
curl -X GET http://localhost:8080/process/status/5ffce937-a5ce-4e37-a65f-4d42ca28a69b
```

**Running Process Response (200):**
```json
{
  "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
  "status": "RUNNING",
  "progress": {
    "total_files": 4,
    "processed_files": 1,
    "percentage": 25
  },
  "started_at": "2025-07-06T17:20:23.175867Z",
  "estimated_completion": "2025-07-06T17:20:23.32092875Z"
}
```

**Completed Process Response (200):**
```json
{
  "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
  "status": "COMPLETED",
  "progress": {
    "total_files": 4,
    "processed_files": 4,
    "percentage": 100
  },
  "started_at": "2025-07-06T17:20:23.175867Z",
  "results": {
    "total_words": 659,
    "total_lines": 58,
    "most_frequent_words": ["the", "and", "system", "-", "to"],
    "files_processed": ["sample1.txt", "sample2.txt", "sample3.txt", "sample4.txt"]
  }
}
```

**Error Response (404):**
```json
{
  "error": "process not found",
  "details": "process with ID 5ffce937-a5ce-4e37-a65f-4d42ca28a69b not found"
}
```

---

### List All Processes

Get a list of all document processing jobs.

**Endpoint:** `GET /process/list`

**Description:** Returns a list of all processing jobs with their current status and progress information.

**Request:** No request body required.

**Response:**
```json
{
  "processes": [
    {
      "process_id": "string",
      "status": "string",
      "progress": {
        "total_files": "integer",
        "processed_files": "integer",
        "percentage": "integer"
      },
      "started_at": "string",
      "estimated_completion": "string"
    }
  ]
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `processes` | array | Array of process summaries |
| `processes[].process_id` | string | Unique identifier for the processing job |
| `processes[].status` | string | Current state of the process |
| `processes[].progress` | object | Progress information |
| `processes[].started_at` | string | ISO 8601 timestamp when processing started |
| `processes[].estimated_completion` | string | ISO 8601 timestamp of estimated completion (only for RUNNING state) |

**Example:**
```bash
curl -X GET http://localhost:8080/process/list
```

**Success Response (200):**
```json
{
  "processes": [
    {
      "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
      "status": "COMPLETED",
      "progress": {
        "total_files": 4,
        "processed_files": 4,
        "percentage": 100
      },
      "started_at": "2025-07-06T17:20:23.175867Z"
    },
    {
      "process_id": "660e8400-e29b-41d4-a716-446655440001",
      "status": "RUNNING",
      "progress": {
        "total_files": 10,
        "processed_files": 3,
        "percentage": 30
      },
      "started_at": "2025-07-06T17:25:00.000000Z",
      "estimated_completion": "2025-07-06T17:26:30.000000Z"
    }
  ]
}
```

---

### Get Analysis Results

Get the aggregated analysis results for a completed document processing job.

**Endpoint:** `GET /process/results/{process_id}`

**Description:** Returns aggregated analysis results and process status information.

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `process_id` | string | Yes | UUID of the process |

**Request:** No request body required.

**Response:**
```json
{
  "process_id": "string",
  "status": "string",
  "progress": {
    "total_files": "integer",
    "processed_files": "integer",
    "percentage": "number"
  },
  "started_at": "string",
  "estimated_completion": "string",
  "results": {
    "total_words": "integer",
    "total_lines": "integer",
    "most_frequent_words": ["string"],
    "files_processed": ["string"]
  }
}
```

**Response Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `process_id` | string | Unique identifier for the processing job |
| `status` | string | Current status of the process |
| `progress.total_files` | integer | Total number of files to process |
| `progress.processed_files` | integer | Number of files already processed |
| `progress.percentage` | number | Completion percentage (0-100) |
| `started_at` | string | ISO 8601 timestamp when processing started |
| `estimated_completion` | string | ISO 8601 timestamp of estimated completion (nullable) |
| `results.total_words` | integer | Total word count across all files |
| `results.total_lines` | integer | Total line count across all files |
| `results.most_frequent_words` | array | Top 5 most frequent words across all files |
| `results.files_processed` | array | List of all processed file names |

**Example:**
```bash
curl -X GET http://localhost:8080/process/results/5ffce937-a5ce-4e37-a65f-4d42ca28a69b
```

**Success Response (200):**
```json
{
  "process_id": "5ffce937-a5ce-4e37-a65f-4d42ca28a69b",
  "status": "COMPLETED",
  "progress": {
    "total_files": 15,
    "processed_files": 15,
    "percentage": 100
  },
  "started_at": "2024-01-15T10:30:00Z",
  "estimated_completion": "2024-01-15T10:32:00Z",
  "results": {
    "total_words": 285740,
    "total_lines": 47875,
    "most_frequent_words": ["shake", "i", "off", "it", "gonna"],
    "files_processed": ["large-sample.txt", "another-sample.txt", "sample.txt", "massive_document_backup.txt", "sample1.txt", "sample2.txt", "ts1.txt", "ts2.txt", "ts3.txt", "ts4.txt", "ts5.txt", "ts6.txt", "ts7.txt", "ts8.txt", "massive_document.txt"]
  }
}
```

**Error Response (500):**
```json
{
  "error": "failed to get results",
  "details": "process not found or not completed"
}
```

## Process States

The system tracks the following process states:

| State | Description | Can Transition To |
|-------|-------------|-------------------|
| `PENDING` | Process created, waiting to start | RUNNING, FAILED |
| `RUNNING` | Currently processing files | COMPLETED, FAILED, STOPPED, PAUSED |
| `PAUSED` | Temporarily paused | RUNNING, FAILED |
| `COMPLETED` | Successfully finished processing | None (terminal state) |
| `FAILED` | Encountered an error during processing | None (terminal state) |
| `STOPPED` | Manually stopped by user | None (terminal state) |

**State Transition Rules:**
- A process starts in `PENDING` state
- When processing begins, it transitions to `RUNNING`
- A running process can be paused, transitioning to `PAUSED`
- A paused process can be resumed, transitioning back to `RUNNING`
- Upon successful completion, it transitions to `COMPLETED`
- If an error occurs, it transitions to `FAILED`
- If manually stopped, it transitions to `STOPPED`
- Terminal states (`COMPLETED`, `FAILED`, `STOPPED`) cannot transition to other states

## Usage Examples

### Complete Workflow Example

1. **Start Processing:**
```bash
# Start processing files in /data folder with batch size 5
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{
    "folder_path": "/data",
    "batch_size": 5
  }'
```

2. **Monitor Progress:**
```bash
# Check status every 10 seconds
while true; do
  curl -s http://localhost:8080/process/status/5ffce937-a5ce-4e37-a65f-4d42ca28a69b | jq .
  sleep 10
done
```

3. **Get Results:**
```bash
# Retrieve analysis results
curl -X GET http://localhost:8080/process/results/5ffce937-a5ce-4e37-a65f-4d42ca28a69b
```

### Error Handling Example

```bash
# Start processing with invalid parameters
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{
    "folder_path": "",
    "batch_size": 0
  }'
```

**Response:**
```json
{
  "error": "invalid request",
  "details": "Key: 'StartProcessRequest.FolderPath' Error:Field validation for 'FolderPath' failed on the 'required' tag"
}
```

### Batch Processing Example

```bash
# Process large number of files with optimal batch size
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{
    "folder_path": "/large-document-collection",
    "batch_size": 20
  }'
```

## Performance Considerations

### Batch Size Optimization
- **Small files (< 1KB)**: Use batch sizes of 10-20
- **Medium files (1KB-100KB)**: Use batch sizes of 5-10
- **Large files (> 100KB)**: Use batch sizes of 2-5

### Worker Pool Configuration
- **Default**: 4 concurrent workers
- **CPU-bound workloads**: Set workers to number of CPU cores
- **I/O-bound workloads**: Can use more workers than CPU cores
- **Memory-constrained environments**: Reduce worker count

### Memory Usage
- System uses streaming I/O to process files efficiently
- Large files are processed without loading entire content into memory
- Frequent words are stored in memory during processing
- Results are persisted to database immediately

### Database Performance
- Connection pooling is configured for optimal performance
- Results are stored with proper indexing
- Process metadata is updated atomically
- Consider database tuning for high-volume processing

### Monitoring and Logging
- Structured logging with correlation IDs for request tracking
- Progress updates are logged at debug level
- Error conditions are logged with full context
- Performance metrics can be extracted from logs

### Rate Limiting
Currently, no rate limiting is implemented. Consider implementing rate limiting for production use:
- Per-client rate limiting
- Per-endpoint rate limiting
- Burst handling for batch operations

### Security Considerations
- Input validation on all endpoints
- Path traversal protection for folder paths
- SQL injection protection through parameterized queries
- Consider implementing authentication for production use 