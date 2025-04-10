#!/bin/bash

# 检查 Docker 和 Docker Compose 是否已安装
if ! command -v docker &> /dev/null; then
    echo "错误: Docker 未安装"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "错误: Docker Compose 未安装"
    exit 1
fi

# 切换到 Docker 目录
cd ../../docker

# 构建并启动服务
echo "构建并启动服务..."
docker-compose up -d --build

echo "部署完成！"
echo "服务端 API: http://localhost:8080"
echo "Web 管理平台: http://localhost"
