#!/bin/bash

# Параметры сборки
APP_NAME="raft-kv-store"
VERSION=$(git describe --tags --always)
BUILD_DIR="bin"
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64")

# Очистка предыдущих сборок
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# Сборка для всех платформ
for PLATFORM in "${PLATFORMS[@]}"; do
  OS=$(echo ${PLATFORM} | cut -d'/' -f1)
  ARCH=$(echo ${PLATFORM} | cut -d'/' -f2)
  OUTPUT="${BUILD_DIR}/${APP_NAME}-${VERSION}-${OS}-${ARCH}"

  echo "Building for ${OS}/${ARCH}..."
  env GOOS=${OS} GOARCH=${ARCH} go build -o ${OUTPUT} ./cmd/server

  if [ $? -ne 0 ]; then
    echo "Build failed for ${OS}/${ARCH}"
    exit 1
  fi

  echo "Build completed: ${OUTPUT}"
done

echo "All builds completed successfully"