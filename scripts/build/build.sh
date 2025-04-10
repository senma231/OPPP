#!/bin/bash

# 设置版本号
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME' -X 'main.CommitHash=$COMMIT_HASH' -s -w"

# 构建服务端
echo "构建服务端..."
cd ../../server
go build -ldflags "$LDFLAGS" -o ../bin/p3-server ./cmd

# 构建客户端
echo "构建客户端..."
cd ../client

# Windows AMD64
echo "构建 Windows AMD64 客户端..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o ../bin/p3-client-windows-amd64.exe ./cmd

# Linux AMD64
echo "构建 Linux AMD64 客户端..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o ../bin/p3-client-linux-amd64 ./cmd

# Linux ARM64
echo "构建 Linux ARM64 客户端..."
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o ../bin/p3-client-linux-arm64 ./cmd

# Linux ARM
echo "构建 Linux ARM 客户端..."
GOOS=linux GOARCH=arm go build -ldflags "$LDFLAGS" -o ../bin/p3-client-linux-arm ./cmd

# macOS AMD64
echo "构建 macOS AMD64 客户端..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o ../bin/p3-client-darwin-amd64 ./cmd

# macOS ARM64
echo "构建 macOS ARM64 客户端..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o ../bin/p3-client-darwin-arm64 ./cmd

echo "构建完成！"
