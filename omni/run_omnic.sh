#!/bin/bash

# Set up the library path for the runtime library
export DYLD_LIBRARY_PATH="$(pwd)/runtime/posix:$DYLD_LIBRARY_PATH"

# Run the omnic compiler with all arguments passed through
./bin/omnic "$@"
