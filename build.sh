#!/bin/bash

# BYD-HASS Build Script
# Builds static binary for Android ARM64

set -e

# Configuration
APP_NAME="byd-hass"
VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DIR="build"
BINARY_NAME="${APP_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building ${APP_NAME} v${VERSION}${NC}"

# Create build directory
mkdir -p ${BUILD_DIR}

# Build for Android ARM64
echo -e "${YELLOW}Building for Android ARM64...${NC}"
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o ${BUILD_DIR}/${BINARY_NAME} \
    cmd/byd-hass/main.go

# Make it executable
chmod +x ${BUILD_DIR}/${BINARY_NAME}

# Show file info
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "Binary: ${BUILD_DIR}/${BINARY_NAME}"
echo -e "Size: $(du -h ${BUILD_DIR}/${BINARY_NAME} | cut -f1)"
echo -e "Target: Linux ARM64 (Android)"

# Usage instructions
echo -e "\n${YELLOW}Usage:${NC}"
echo "  # Copy to Android device"
echo "  adb push ${BUILD_DIR}/${BINARY_NAME} /data/data/com.termux/files/usr/bin/"
echo ""
echo "  # Run on device"
echo "  ${BINARY_NAME} --help"
echo "  ${BINARY_NAME} --mqtt-url ws://broker:9001/mqtt --abrp-api-key YOUR_KEY --abrp-vehicle-key YOUR_VEHICLE" 