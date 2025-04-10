.PHONY: all server client clean

all: server client

server:
	@echo "Building server..."
	cd server && go build -o ../bin/p3-server ./cmd

client:
	@echo "Building client..."
	cd client && go build -o ../bin/p3-client ./cmd

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

clean:
	@echo "Cleaning..."
	rm -rf bin/

setup:
	@echo "Creating directories..."
	mkdir -p bin
	@echo "Installing dependencies..."
	cd server && go mod tidy
	cd client && go mod tidy
