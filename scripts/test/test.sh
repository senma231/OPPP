#!/bin/bash

# 测试服务端
echo "测试服务端..."
cd ../../server
go test -v ./...

# 测试客户端
echo "测试客户端..."
cd ../client
go test -v ./...
