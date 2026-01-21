.PHONY: build clean test release

# Binary name
BINARY=note-sync

# Build directory
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/note-sync

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

# Cross-compile for multiple platforms
release: clean
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/note-sync
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/note-sync
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/note-sync
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/note-sync
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/note-sync

# Install locally
install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/

# Download dependencies
deps:
	go mod download
	go mod tidy
