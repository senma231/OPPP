.PHONY: all server client clean test lint docker release

# 版本信息
VERSION := 0.1.0
BUILD := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"

all: server client

server:
	@echo "Building server..."
	cd server && go build $(LDFLAGS) -o ../bin/p3-server .

client:
	@echo "Building client..."
	cd client && go build $(LDFLAGS) -o ../bin/p3-client ./cmd

client-windows:
	@echo "Building Windows client..."
	cd client && GOOS=windows GOARCH=amd64 go build -o ../bin/p3-client-windows-amd64.exe ./cmd

client-linux:
	@echo "Building Linux client..."
	cd client && GOOS=linux GOARCH=amd64 go build -o ../bin/p3-client-linux-amd64 ./cmd

client-linux-arm:
	@echo "Building Linux ARM client..."
	cd client && GOOS=linux GOARCH=arm go build -o ../bin/p3-client-linux-arm ./cmd

client-linux-arm64:
	@echo "Building Linux ARM64 client..."
	cd client && GOOS=linux GOARCH=arm64 go build -o ../bin/p3-client-linux-arm64 ./cmd

client-macos:
	@echo "Building macOS client..."
	cd client && GOOS=darwin GOARCH=amd64 go build -o ../bin/p3-client-darwin-amd64 ./cmd

client-macos-arm:
	@echo "Building macOS ARM client..."
	cd client && GOOS=darwin GOARCH=arm64 go build -o ../bin/p3-client-darwin-arm64 ./cmd

all-clients: client-windows client-linux client-linux-arm client-linux-arm64 client-macos client-macos-arm

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 运行代码检查
lint:
	@echo "Running linters..."
	go vet ./...

# 构建 Docker 镜像
docker: docker-server docker-client

# 构建服务端 Docker 镜像
docker-server:
	@echo "Building server Docker image..."
	docker build -t senma231/p3-server:$(VERSION) -f docker/server/Dockerfile .
	docker tag senma231/p3-server:$(VERSION) senma231/p3-server:latest

# 构建客户端 Docker 镜像
docker-client:
	@echo "Building client Docker image..."
	docker build -t senma231/p3-client:$(VERSION) -f docker/client/Dockerfile .
	docker tag senma231/p3-client:$(VERSION) senma231/p3-client:latest

# 创建发布包
release: all-clients
	@echo "Creating release packages..."
	mkdir -p release
	# Windows
	zip -j release/p3-client-windows-amd64-$(VERSION).zip bin/p3-client-windows-amd64.exe client/config.example.yaml
	# Linux
	tar -czf release/p3-client-linux-amd64-$(VERSION).tar.gz -C bin p3-client-linux-amd64 -C ../client config.example.yaml
	tar -czf release/p3-client-linux-arm-$(VERSION).tar.gz -C bin p3-client-linux-arm -C ../client config.example.yaml
	tar -czf release/p3-client-linux-arm64-$(VERSION).tar.gz -C bin p3-client-linux-arm64 -C ../client config.example.yaml
	# macOS
	tar -czf release/p3-client-darwin-amd64-$(VERSION).tar.gz -C bin p3-client-darwin-amd64 -C ../client config.example.yaml
	tar -czf release/p3-client-darwin-arm64-$(VERSION).tar.gz -C bin p3-client-darwin-arm64 -C ../client config.example.yaml

# 初始化数据库
init-db:
	@echo "Initializing database..."
	bin/p3-server -init-db

# 运行服务端
run-server:
	@echo "Running server..."
	bin/p3-server -config server/config.yaml

# 运行客户端
run-client:
	@echo "Running client..."
	bin/p3-client -config client/config.yaml

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf release/

setup:
	@echo "Creating directories..."
	mkdir -p bin
	@echo "Installing dependencies..."
	go mod tidy
