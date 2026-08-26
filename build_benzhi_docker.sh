#!/usr/bin/env bash
set -euo pipefail

IMAGE_NAME="${IMAGE_NAME:-epochclock-benzhi}"

docker build -f benzhi.Dockerfile -t "${IMAGE_NAME}" .

echo "image ${IMAGE_NAME} built successfully"
