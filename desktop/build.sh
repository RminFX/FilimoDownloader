#!/bin/sh
set -e
cd "$(dirname "$0")"
export PATH="$PATH:$HOME/go/bin"

case "${1:-mac-arm}" in
  darwin|mac-arm)
    wails build -platform darwin/arm64
    ;;
  mac-intel)
    wails build -platform darwin/amd64
    ;;
  windows)
    wails build -platform windows/amd64
    ;;
  linux)
    wails build -platform linux/amd64
    ;;
  *)
    echo "Usage: $0 [mac-arm|mac-intel|windows|linux]"
    exit 1
    ;;
esac
