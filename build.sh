#!/usr/bin/env bash

set -e 

echo "Detecting system arch..."

if [[ "${OSTYPE}" == "darwin"* ]]; then
  # We're on Mac, check for apple silicon (arm64)
  if [[ $(uname -m) == "arm64" ]]; then
    echo "Apple silicon mac detected"
    export CGO_ENABLED=1
    export GOARCH=arm64
  else
    export CGO_ENABLED=1
    export GOARCH=amd64
  if
elif [[ "${OSTYPE}" == "linux-gnu"* ]]; then
  echo "Linux detected"
  export CGO_ENABLED=1
else 
  echo "Other OS (windows?) detected"
  export CGO_ENABLED=1
fi

echo -e "\nBuilding CHIP-8 emulator..."

go build -o chip8-emulator -v

if [ $? -eq 0 ]; then
  echo "Build successful"
  echo "Output: ./chip8-emulator"
  exit 0
else
  echo "Build failed"
  exit 1
fi

