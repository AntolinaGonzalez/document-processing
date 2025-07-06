# 🚀 Document Processing System Demo

## Quick Start Demo

Once you have Docker Compose installed, follow these steps:

### 1. Start the System
```bash
docker compose up -d
```

This will start:
- PostgreSQL database (port 5432)
- Document Processing System (port 8080)

### 2. Run the Demo Test
```bash
./demo-test.sh
```

### 3. What the Demo Does

The demo script will:

1. **Health Check** - Verify the service is running
2. **Start Processing** - Begin processing the 2 sample text files
3. **Monitor Progress** - Check status every 3 seconds
4. **Get Results** - Retrieve the analysis results
5. **List Processes** - Show all processing jobs

### 4. Expected Results

You should see:
- ✅ Service health check passes
- ✅ Process starts with a unique ID
- ✅ Progress updates from 0% to 100%
- ✅ Analysis results for both text files
- ✅ Process list showing completed job

### 5. Sample API Responses

**Health Check:**
```json
{"status":"ok","timestamp":1704110400}
```

**Start Process:**
```json
{"process_id":"550e8400-e29b-41d4-a716-446655440000","status":"PENDING"}
```

**Process Status:**
```json
{
  "process_id":"550e8400-e29b-41d4-a716-446655440000",
  "status":"RUNNING",
  "progress":{
    "total_files":15,
    "processed_files":7,
    "percentage":46.67
  },
  "started_at":"2024-01-15T10:30:00Z",
  "estimated_completion":"2024-01-15T10:32:00Z",
  "results":{
    "total_words":12500,
    "total_lines":850,
    "most_frequent_words":["the","and","to","of","a"],
    "files_processed":["file1.txt","file2.txt","file3.txt"]
  }
}
```

**Analysis Results:**
```json
{
  "process_id":"550e8400-e29b-41d4-a716-446655440000",
  "status":"COMPLETED",
  "progress":{
    "total_files":15,
    "processed_files":15,
    "percentage":100
  },
  "started_at":"2024-01-15T10:30:00Z",
  "estimated_completion":"2024-01-15T10:32:00Z",
  "results":{
    "total_words":285740,
    "total_lines":47875,
    "most_frequent_words":["shake","i","off","it","gonna"],
    "files_processed":["large-sample.txt","another-sample.txt","sample.txt","massive_document_backup.txt","sample1.txt","sample2.txt","ts1.txt","ts2.txt","ts3.txt","ts4.txt","ts5.txt","ts6.txt","ts7.txt","ts8.txt","massive_document.txt"]
  }
}
```

### 6. Clean Up
```bash
docker compose down
```

## Manual Testing

You can also test individual endpoints manually:

```bash
# Health check
curl http://localhost:8080/health

# Start processing
curl -X POST http://localhost:8080/process/start \
  -H "Content-Type: application/json" \
  -d '{"folder_path":"/data","batch_size":2}'

# Check status (replace PROCESS_ID with actual ID)
curl http://localhost:8080/process/status/PROCESS_ID

# Get results (replace PROCESS_ID with actual ID)
curl http://localhost:8080/process/results/PROCESS_ID
```

## Troubleshooting

- **Port already in use**: Make sure ports 8080 and 5432 are available
- **Database connection issues**: Wait a few seconds for PostgreSQL to start
- **Permission denied**: Make sure demo-test.sh is executable (`chmod +x demo-test.sh`)

## What's Being Processed

The demo processes these files from `test-data/`:
- `sample.txt` - Document processing system overview
- `another-sample.txt` - Concurrent processing in Go

The system will analyze each file and extract:
- Word count, line count, character count
- Most frequent words (top 5)
- Content summary (first 2 sentences)

**Enhanced Features:**
- 📊 Real-time progress tracking with percentage
- ⏰ Estimated completion time calculation
- 📈 Aggregated results summary across all files
- 📝 Most frequent words across all processed files
- 📁 List of all processed files 

3️⃣  Monitoring process status with enhanced progress tracking...
   Check 1:
   📈 Progress: 1/2 files (50%)
   ⏰ Estimated completion: 2024-01-15T10:31:30Z
   
   Check 2:
   ✅ Process completed!
   
   📊 Final Results Summary:
   - Total files processed: 2
   - Total words analyzed: 450
   - Total lines analyzed: 33
   - Most frequent words: the, system, processing, files, and
   - Files processed: sample.txt, another-sample.txt 