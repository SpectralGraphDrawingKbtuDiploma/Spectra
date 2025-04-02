#!/bin/bash

# Function to log errors
log_error() {
    echo "[ERROR] $(date +'%Y-%m-%d %H:%M:%S') - $1" >> "$2/error.txt"
    echo "[ERROR] $1" >&2
}

# Set -e ensures the script exits if any command fails
set -e

echo "Starting processing pipeline..."

# Compile C++ code with optimization
echo "Compiling C++ code..."
if ! g++ script.cpp -I/usr/local/include/eigen3 -o spectral_embed -O2; then
    log_error "Failed to compile C++ code" "$2"
    exit 1
fi

# Run executable with arguments
echo "Running spectral embedding..."
if ! ./spectral_embed "$1" 1 1 3 "$2"; then
    log_error "Failed to run spectral embedding" "$2"
    exit 1
fi

# Compile and run C drawing code
echo "Compiling drawing code..."
if ! gcc -std=c99 -O2 -o draw draw.c -lm; then
    log_error "Failed to compile drawing code" "$2"
    exit 1
fi

echo "Generating visualization..."
if ! ./draw "$1" "$2/embedding.txt" "$2/out.png"; then
    log_error "Failed to generate visualization" "$2"
    exit 1
fi

# Upload results to S3/MinIO
echo "Uploading results to storage..."
if ! /app/venv/bin/python ./upload_to_s3.py --local-path "$2" --s3-directory "$3"; then
    log_error "Failed to upload files to storage" "$2"
    exit 1
fi

echo "Processing pipeline completed successfully!"
exit 0
