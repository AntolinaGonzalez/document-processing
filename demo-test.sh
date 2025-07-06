#!/bin/bash

# Document Processing System Demo Test
# This script demonstrates the API functionality with enhanced responses

echo "🚀 Document Processing System Demo Test"
echo "========================================"

# Configuration
API_BASE="http://localhost:8080"
TEST_FOLDER="/data"  # This would be mounted in Docker
BATCH_SIZE=2

echo ""
echo "1️⃣  Checking if the service is running..."
echo "   GET $API_BASE/health"
echo ""

# Health check
HEALTH_RESPONSE=$(curl -s "$API_BASE/health")
if [ $? -eq 0 ]; then
    echo "✅ Service is running!"
    echo "   Response: $HEALTH_RESPONSE"
else
    echo "❌ Service is not running. Please start the service first."
    echo "   You can start it with: docker compose up -d"
    exit 1
fi

echo ""
echo "2️⃣  Starting document processing..."
echo "   POST $API_BASE/process/start"
echo "   Body: {\"folder_path\": \"$TEST_FOLDER\", \"batch_size\": $BATCH_SIZE}"
echo ""

# Start processing
START_RESPONSE=$(curl -s -X POST "$API_BASE/process/start" \
    -H "Content-Type: application/json" \
    -d "{\"folder_path\": \"$TEST_FOLDER\", \"batch_size\": $BATCH_SIZE}")

if [ $? -eq 0 ]; then
    echo "✅ Process started successfully!"
    echo "   Response: $START_RESPONSE"
    
    # Extract process ID from response
    PROCESS_ID=$(echo $START_RESPONSE | grep -o '"process_id":"[^"]*"' | cut -d'"' -f4)
    echo "   Process ID: $PROCESS_ID"
else
    echo "❌ Failed to start process"
    echo "   Response: $START_RESPONSE"
    exit 1
fi

echo ""
echo "3️⃣  Monitoring process status with enhanced progress tracking..."
echo "   GET $API_BASE/process/status/$PROCESS_ID"
echo ""

# Monitor status for a few iterations
for i in {1..5}; do
    echo "   Check $i:"
    STATUS_RESPONSE=$(curl -s "$API_BASE/process/status/$PROCESS_ID")
    if [ $? -eq 0 ]; then
        echo "   Response: $STATUS_RESPONSE"
        
        # Check if process is completed
        if echo "$STATUS_RESPONSE" | grep -q '"status":"COMPLETED"'; then
            echo "   ✅ Process completed!"
            echo ""
            echo "   📊 Final Results Summary:"
            echo "   - Total files processed: $(echo $STATUS_RESPONSE | grep -o '"total_files":[0-9]*' | cut -d':' -f2)"
            echo "   - Total words analyzed: $(echo $STATUS_RESPONSE | grep -o '"total_words":[0-9]*' | cut -d':' -f2)"
            echo "   - Total lines analyzed: $(echo $STATUS_RESPONSE | grep -o '"total_lines":[0-9]*' | cut -d':' -f2)"
            echo "   - Most frequent words: $(echo $STATUS_RESPONSE | grep -o '"most_frequent_words":\[[^]]*\]' | sed 's/.*\[\(.*\)\].*/\1/')"
            echo "   - Files processed: $(echo $STATUS_RESPONSE | grep -o '"files_processed":\[[^]]*\]' | sed 's/.*\[\(.*\)\].*/\1/')"
            break
        elif echo "$STATUS_RESPONSE" | grep -q '"status":"FAILED"'; then
            echo "   ❌ Process failed!"
            break
        else
            # Show progress info
            TOTAL_FILES=$(echo $STATUS_RESPONSE | grep -o '"total_files":[0-9]*' | cut -d':' -f2)
            PROCESSED_FILES=$(echo $STATUS_RESPONSE | grep -o '"processed_files":[0-9]*' | cut -d':' -f2)
            PERCENTAGE=$(echo $STATUS_RESPONSE | grep -o '"percentage":[0-9]*' | cut -d':' -f2)
            ESTIMATED_COMPLETION=$(echo $STATUS_RESPONSE | grep -o '"estimated_completion":"[^"]*"' | cut -d'"' -f4)
            
            echo "   📈 Progress: $PROCESSED_FILES/$TOTAL_FILES files ($PERCENTAGE%)"
            if [ ! -z "$ESTIMATED_COMPLETION" ]; then
                echo "   ⏰ Estimated completion: $ESTIMATED_COMPLETION"
            fi
        fi
    else
        echo "   ❌ Failed to get status"
    fi
    
    if [ $i -lt 5 ]; then
        echo "   Waiting 3 seconds..."
        sleep 3
    fi
done

echo ""
echo "4️⃣  Getting detailed analysis results..."
echo "   GET $API_BASE/process/results/$PROCESS_ID"
echo ""

# Get results
RESULTS_RESPONSE=$(curl -s "$API_BASE/process/results/$PROCESS_ID")
if [ $? -eq 0 ]; then
    echo "✅ Results retrieved successfully!"
    echo "   Response: $RESULTS_RESPONSE"
else
    echo "❌ Failed to get results"
    echo "   Response: $RESULTS_RESPONSE"
fi

echo ""
echo "5️⃣  Listing all processes with enhanced progress info..."
echo "   GET $API_BASE/process/list"
echo ""

# List all processes
LIST_RESPONSE=$(curl -s "$API_BASE/process/list")
if [ $? -eq 0 ]; then
    echo "✅ Process list retrieved!"
    echo "   Response: $LIST_RESPONSE"
else
    echo "❌ Failed to get process list"
    echo "   Response: $LIST_RESPONSE"
fi

echo ""
echo "🎉 Demo test completed!"
echo "======================"
echo ""
echo "Expected Results Summary:"
echo "- Health check should return status 'ok'"
echo "- Process start should return a process_id"
echo "- Status should show detailed progress with estimated completion"
echo "- Results should contain analysis of 2 text files"
echo "- Process list should show the completed process with progress info"
echo ""
echo "Enhanced Features Demonstrated:"
echo "- 📊 Real-time progress tracking with percentage"
echo "- ⏰ Estimated completion time calculation"
echo "- 📈 Aggregated results summary"
echo "- 📝 Most frequent words across all files"
echo "- 📁 List of processed files"
echo ""
echo "To run this demo with Docker:"
echo "1. docker compose up -d"
echo "2. ./demo-test.sh"
echo "3. docker compose down" 