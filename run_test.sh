#!/bin/bash

# Simple script to run the watermark remover test inside a Docker container
# to avoid local dependency issues with gocv/opencv.

URL=$1
if [ -z "$URL" ]; then
    echo "Usage: ./run_test.sh <IMAGE_URL>"
    exit 1
fi

# 1. Build the tester image (if it doesn't exist or --build is passed)
if [[ "$(docker images -q watermark-tester 2> /dev/null)" == "" ]] || [[ "$2" == "--build" ]]; then
    echo "--- Building Tester Image ---"
    docker build -t watermark-tester -f Dockerfile .
else
    echo "--- Using existing watermark-tester image (use --build to rebuild) ---"
fi

# 2. Run the tester utility inside a temporary container
# We mount the output directory to get the result back
mkdir -p output
docker run --rm \
    -v "$(pwd)/output:/app/output" \
    watermark-tester \
    ./tester -url "$URL"

echo "--- Test Complete! Check the 'output/' directory. ---"
ls -lh output/
