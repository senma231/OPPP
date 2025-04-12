#!/bin/bash

# P3 发布脚本
# 用于创建发布包和 Docker 镜像

set -e

# 版本信息
VERSION=$(grep "VERSION :=" Makefile | cut -d "=" -f 2 | tr -d " ")
if [ -z "$VERSION" ]; then
    VERSION="0.1.0"
fi

# 创建发布目录
mkdir -p release

echo "开始创建 P3 v$VERSION 发布包..."

# 构建所有平台
echo "构建所有平台..."
make build-all

# 创建发布包
echo "创建发布包..."

# Linux AMD64
echo "打包 Linux AMD64..."
tar -czf release/p3-server-linux-amd64-$VERSION.tar.gz -C bin p3-server-linux-amd64 -C ../server config.example.yaml
tar -czf release/p3-client-linux-amd64-$VERSION.tar.gz -C bin p3-client-linux-amd64 -C ../client config.example.yaml

# Linux ARM64
echo "打包 Linux ARM64..."
tar -czf release/p3-server-linux-arm64-$VERSION.tar.gz -C bin p3-server-linux-arm64 -C ../server config.example.yaml
tar -czf release/p3-client-linux-arm64-$VERSION.tar.gz -C bin p3-client-linux-arm64 -C ../client config.example.yaml

# macOS AMD64
echo "打包 macOS AMD64..."
tar -czf release/p3-server-darwin-amd64-$VERSION.tar.gz -C bin p3-server-darwin-amd64 -C ../server config.example.yaml
tar -czf release/p3-client-darwin-amd64-$VERSION.tar.gz -C bin p3-client-darwin-amd64 -C ../client config.example.yaml

# macOS ARM64
echo "打包 macOS ARM64..."
tar -czf release/p3-server-darwin-arm64-$VERSION.tar.gz -C bin p3-server-darwin-arm64 -C ../server config.example.yaml
tar -czf release/p3-client-darwin-arm64-$VERSION.tar.gz -C bin p3-client-darwin-arm64 -C ../client config.example.yaml

# Windows AMD64
echo "打包 Windows AMD64..."
zip -j release/p3-server-windows-amd64-$VERSION.zip bin/p3-server-windows-amd64.exe server/config.example.yaml
zip -j release/p3-client-windows-amd64-$VERSION.zip bin/p3-client-windows-amd64.exe client/config.example.yaml

# 构建 Docker 镜像
echo "构建 Docker 镜像..."
make docker

# 创建 SHA256 校验和
echo "创建校验和..."
cd release
sha256sum * > SHA256SUMS
cd ..

echo "发布包创建完成！"
echo "发布包位于 release/ 目录"
echo "Docker 镜像: senma231/p3-server:$VERSION, senma231/p3-client:$VERSION"

# 显示发布说明
cat << EOF

P3 v$VERSION 发布说明
=====================

发布包:
- Linux AMD64: p3-server-linux-amd64-$VERSION.tar.gz, p3-client-linux-amd64-$VERSION.tar.gz
- Linux ARM64: p3-server-linux-arm64-$VERSION.tar.gz, p3-client-linux-arm64-$VERSION.tar.gz
- macOS AMD64: p3-server-darwin-amd64-$VERSION.tar.gz, p3-client-darwin-amd64-$VERSION.tar.gz
- macOS ARM64: p3-server-darwin-arm64-$VERSION.tar.gz, p3-client-darwin-arm64-$VERSION.tar.gz
- Windows AMD64: p3-server-windows-amd64-$VERSION.zip, p3-client-windows-amd64-$VERSION.zip

Docker 镜像:
- 服务端: senma231/p3-server:$VERSION, senma231/p3-server:latest
- 客户端: senma231/p3-client:$VERSION, senma231/p3-client:latest

SHA256 校验和:
- SHA256SUMS

发布说明:
- 初始版本发布
- 支持 P2P 连接和端口转发
- 支持 NAT 穿透和 UPnP
- 支持 TLS 加密和 JWT 认证
- 支持 Web 管理界面

EOF
