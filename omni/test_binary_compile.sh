#!/bin/bash
# OmniLang C Compilation Script
echo "Compiling OmniLang program..."

# Find the runtime directory
RUNTIME_DIR="$(dirname "$0")/../runtime"
if [ ! -d "$RUNTIME_DIR" ]; then
    echo "Error: Runtime directory not found at $RUNTIME_DIR"
    exit 1
fi

# Compile with gcc
gcc -o "test_binary" "test_binary.c" "$RUNTIME_DIR/omni_rt.c" -I"$RUNTIME_DIR"

if [ $? -eq 0 ]; then
    echo "Compilation successful: test_binary"
else
    echo "Compilation failed"
    exit 1
fi
