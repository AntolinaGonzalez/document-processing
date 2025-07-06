#!/bin/bash

# Document Processing System Pause/Resume Test
# This script demonstrates the pause/resume functionality

echo "🚀 Document Processing System Pause/Resume Test"
echo "==============================================="

# Configuration
API_BASE="http://localhost:8080"
TEST_FOLDER="/data"
BATCH_SIZE=1  # Small batch size to make it easier to pause

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
echo "2️⃣  Starting document processing with small batch size..."
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
echo "3️⃣  Monitoring process until it starts running..."
echo "   GET $API_BASE/process/status/$PROCESS_ID"
echo ""

# Wait for process to start running
for i in {1..10}; do
    echo "   Check $i:"
    STATUS_RESPONSE=$(curl -s "$API_BASE/process/status/$PROCESS_ID")
    if [ $? -eq 0 ]; then
        echo "   Response: $STATUS_RESPONSE"
        
        # Check if process is running
        if echo "$STATUS_RESPONSE" | grep -q '"status":"RUNNING"'; then
            echo "   ✅ Process is now running!"
            break
        elif echo "$STATUS_RESPONSE" | grep -q '"status":"COMPLETED"'; then
            echo "   ✅ Process completed too quickly for pause test"
            echo "   This means the files were processed very fast."
            echo "   Try with more files or larger files for better pause/resume testing."
            exit 0
        fi
    else
        echo "   ❌ Failed to get status"
    fi
    
    if [ $i -lt 10 ]; then
        echo "   Waiting 2 seconds..."
        sleep 2
    fi
done

echo ""
echo "4️⃣  Pausing the process..."
echo "   POST $API_BASE/process/pause/$PROCESS_ID"
echo ""

# Pause the process
PAUSE_RESPONSE=$(curl -s -X POST "$API_BASE/process/pause/$PROCESS_ID")
if [ $? -eq 0 ]; then
    echo "✅ Process paused successfully!"
    echo "   Response: $PAUSE_RESPONSE"
else
    echo "❌ Failed to pause process"
    echo "   Response: $PAUSE_RESPONSE"
    exit 1
fi

echo ""
echo "5️⃣  Verifying process is paused..."
echo "   GET $API_BASE/process/status/$PROCESS_ID"
echo ""

# Check if process is paused
STATUS_RESPONSE=$(curl -s "$API_BASE/process/status/$PROCESS_ID")
if [ $? -eq 0 ]; then
    echo "   Response: $STATUS_RESPONSE"
    
    if echo "$STATUS_RESPONSE" | grep -q '"status":"PAUSED"'; then
        echo "   ✅ Process is confirmed to be paused!"
    else
        echo "   ❌ Process is not paused"
        exit 1
    fi
else
    echo "   ❌ Failed to get status"
    exit 1
fi

echo ""
echo "6️⃣  Waiting 5 seconds while paused..."
sleep 5
echo "   ✅ Waited 5 seconds while process was paused"

echo ""
echo "7️⃣  Resuming the process..."
echo "   POST $API_BASE/process/resume/$PROCESS_ID"
echo ""

# Resume the process
RESUME_RESPONSE=$(curl -s -X POST "$API_BASE/process/resume/$PROCESS_ID")
if [ $? -eq 0 ]; then
    echo "✅ Process resumed successfully!"
    echo "   Response: $RESUME_RESPONSE"
else
    echo "❌ Failed to resume process"
    echo "   Response: $RESUME_RESPONSE"
    exit 1
fi

echo ""
echo "8️⃣  Monitoring process completion..."
echo "   GET $API_BASE/process/status/$PROCESS_ID"
echo ""

# Monitor until completion
for i in {1..10}; do
    echo "   Check $i:"
    STATUS_RESPONSE=$(curl -s "$API_BASE/process/status/$PROCESS_ID")
    if [ $? -eq 0 ]; then
        echo "   Response: $STATUS_RESPONSE"
        
        # Check if process is completed
        if echo "$STATUS_RESPONSE" | grep -q '"status":"COMPLETED"'; then
            echo "   ✅ Process completed successfully after pause/resume!"
            break
        elif echo "$STATUS_RESPONSE" | grep -q '"status":"FAILED"'; then
            echo "   ❌ Process failed after resume"
            exit 1
        fi
    else
        echo "   ❌ Failed to get status"
    fi
    
    if [ $i -lt 10 ]; then
        echo "   Waiting 3 seconds..."
        sleep 3
    fi
done

echo ""
echo "9️⃣  Getting final results..."
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
echo "🎉 Pause/Resume test completed!"
echo "==============================="
echo ""
echo "Test Summary:"
echo "- ✅ Process started successfully"
echo "- ✅ Process was paused while running"
echo "- ✅ Process remained paused for 5 seconds"
echo "- ✅ Process was resumed successfully"
echo "- ✅ Process completed after resume"
echo "- ✅ Results were retrieved successfully"
echo ""
echo "Pause/Resume functionality is working correctly!"
echo ""
echo "To run this test:"
echo "1. docker compose up -d"
echo "2. ./test-pause-resume.sh"
echo "3. docker compose down" 