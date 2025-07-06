# Document Processing System API Documentation

## Overview
The Document Processing System provides a REST API for asynchronously processing text files and extracting analysis results. All endpoints return JSON responses and use standard HTTP status codes.

**Base URL**: `http://localhost:8080`

## Authentication
Currently, the API does not require authentication. In production, consider implementing API keys or JWT tokens.

## Error Handling
All endpoints return consistent error responses in the following format:
```json
{
  "error": "Error description",
  "details": "Additional error details"
}
```

Common HTTP status codes:
- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Endpoints

### 1. Health Check
Check if the service is running and healthy.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "ok",
  "timestamp": 1704110400
}
```

**Example**:
```bash
curl http://localhost:8080/health
```

### 2. Start Document Processing
Start a new document processing job for files in a specified folder.

**Endpoint**: `POST /process/start`

**Request Body**:
```json
{
  "folder_path": "/path/to/text/files",
  "batch_size": 10
}
```

**Parameters**:
- `folder_path` (string, required): Path to folder containing .txt files
- `batch_size` (integer, required, min: 1): Number of files to process in each batch

**Response**:
```json
{
  "process_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "PENDING"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{
    "folder_path": "/data",
    "batch_size": 5
  }'
```

### 3. Stop Document Processing
Stop a running document processing job.

**Endpoint**: `POST /process/stop/{process_id}`

**Parameters**:
- `process_id` (string, required): UUID of the process to stop

**Response**:
```json
{
  "process_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "STOPPED",
  "message": "Process stopped"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/process/stop/550e8400-e29b-41d4-a716-446655440000
```

### 4. Get Process Status
Get the current status and progress of a document processing job.

**Endpoint**: `GET /process/status/{process_id}`

**Parameters**:
- `process_id` (string, required): UUID of the process

**Response**:
```json
{
  "process_id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "RUNNING",
  "progress": 45.5,
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": null,
  "error_detail": null
}
```

**Example**:
```bash
curl http://localhost:8080/process/status/550e8400-e29b-41d4-a716-446655440000
```

### 5. List All Processes
Get a list of all document processing jobs.

**Endpoint**: `GET /process/list`

**Response**:
```json
{
  "processes": [
    {
      "process_id": "550e8400-e29b-41d4-a716-446655440000",
      "state": "COMPLETED",
      "progress": 100.0,
      "start_time": "2024-01-01T10:00:00Z",
      "end_time": "2024-01-01T10:05:00Z"
    },
    {
      "process_id": "660e8400-e29b-41d4-a716-446655440001",
      "state": "RUNNING",
      "progress": 25.0,
      "start_time": "2024-01-01T10:10:00Z",
      "end_time": null
    }
  ]
}
```

**Example**:
```bash
curl http://localhost:8080/process/list
```

### 6. Get Analysis Results
Get the analysis results for a completed document processing job.

**Endpoint**: `GET /process/results/{process_id}`

**Parameters**:
- `process_id` (string, required): UUID of the process

**Response**:
```json
{
  "process_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "file_name": "document.txt",
      "word_count": 1500,
      "line_count": 75,
      "char_count": 8500,
      "frequent_words": {
        "the": 45,
        "and": 32,
        "to": 28,
        "of": 25,
        "in": 20
      },
      "summary": "This document discusses the implementation of a document processing system. The system provides comprehensive analysis capabilities for text files."
    },
    {
      "file_name": "another-document.txt",
      "word_count": 800,
      "line_count": 40,
      "char_count": 4500,
      "frequent_words": {
        "system": 15,
        "processing": 12,
        "files": 10,
        "analysis": 8,
        "text": 6
      },
      "summary": "The document processing system analyzes text files efficiently. It provides detailed statistics and insights."
    }
  ]
}
```

**Example**:
```bash
curl http://localhost:8080/process/results/550e8400-e29b-41d4-a716-446655440000
```

## Process States

The system tracks the following process states:

| State | Description |
|-------|-------------|
| `PENDING` | Process created, waiting to start |
| `RUNNING` | Currently processing files |
| `PAUSED` | Temporarily paused (not implemented) |
| `COMPLETED` | Successfully finished processing |
| `FAILED` | Encountered an error during processing |
| `STOPPED` | Manually stopped by user |

## Usage Examples

### Complete Workflow Example

1. **Start Processing**:
```bash
# Start processing files in /data folder with batch size 5
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{"folder_path": "/data", "batch_size": 5}'
```

2. **Monitor Progress**:
```bash
# Check status every 10 seconds
while true; do
  curl http://localhost:8080/process/status/550e8400-e29b-41d4-a716-446655440000
  sleep 10
done
```

3. **Get Results**:
```bash
# Retrieve analysis results
curl http://localhost:8080/process/results/550e8400-e29b-41d4-a716-446655440000
```

### Error Handling Example

```bash
# Start processing with invalid parameters
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{"folder_path": "", "batch_size": 0}'
```

**Response**:
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
  -d '{"folder_path": "/large-document-collection", "batch_size": 20}'
```

## Performance Considerations

- **Batch Size**: Optimal batch size depends on file sizes and system resources
- **Worker Count**: Default is 4 workers, can be configured in the application
- **Memory Usage**: System uses streaming I/O to handle large files efficiently
- **Database**: Results are stored in PostgreSQL with proper indexing

## Rate Limiting

Currently, no rate limiting is implemented. Consider implementing rate limiting for production use.

## Monitoring

The system provides structured logging with the following log levels:
- `debug`: Detailed debugging information
- `info`: General operational information  
- `warn`: Warning messages
- `error`: Error conditions

Logs include correlation IDs (process IDs) for tracking requests across the system. 