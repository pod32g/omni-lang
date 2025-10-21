#!/bin/bash

# Safe runner script with timeout protection for loop testing
# Usage: ./safe_run.sh <command> [timeout_seconds]

COMMAND="$1"
TIMEOUT="${2:-5}"  # Default 5 seconds timeout

echo "Running: $COMMAND"
echo "Timeout: ${TIMEOUT}s"

# Run the command in background
eval "$COMMAND" &
PID=$!

# Wait for the specified timeout
sleep $TIMEOUT

# Check if process is still running
if kill -0 $PID 2>/dev/null; then
    echo "Process still running after ${TIMEOUT}s - killing it"
    kill $PID 2>/dev/null
    sleep 1
    # Force kill if still running
    if kill -0 $PID 2>/dev/null; then
        kill -9 $PID 2>/dev/null
    fi
    echo "Process killed"
    exit 1
else
    echo "Process completed normally"
    wait $PID
    echo "Exit code: $?"
fi
