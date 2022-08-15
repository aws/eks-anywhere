#!/bin/bash

set -ex

dir=$(realpath "$(dirname "$0")")
cd "$dir"

TIMESTAMP=$(date +'%Y-%m-%d')
IMAGE_NAME='tiltdev/restart-helper'
IMAGE_WITH_TAG=$IMAGE_NAME:$TIMESTAMP

# build binary for tilt-restart-wrapper
env GOOS=linux GOARCH=amd64 go build tilt-restart-wrapper.go

# build Docker image with static binaries of:
# - tilt-restart-wrapper (compiled above)
# - entr (dependency of tilt-restart-wrapper)
docker buildx build \
  --push --no-cache \
  --platform linux/arm/v7,linux/arm64/v8,linux/amd64 \
  -t "$IMAGE_NAME" .
docker buildx build \
  --push --no-cache \
  --platform linux/arm/v7,linux/arm64/v8,linux/amd64 \
  -t "$IMAGE_WITH_TAG" .

echo "Successfully built and pushed $IMAGE_WITH_TAG"



