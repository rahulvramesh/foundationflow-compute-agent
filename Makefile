# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Name of your Go package
BINARY_NAME=server-monitor-agent
BINARY_UNIX=$(BINARY_NAME)_unix

# Architectures to build for
ARCHITECTURES=amd64 386 arm arm64

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -rf build

run:
	$(GOBUILD) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

deps:
	$(GOGET) github.com/shirou/gopsutil/v3
	$(GOGET) github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml
	$(GOGET) github.com/mattn/go-sqlite3

# Cross compilation
build-linux:
	mkdir -p build
	$(foreach ARCH,$(ARCHITECTURES),\
		echo "Building for linux/$(ARCH)..." && \
		GOOS=linux GOARCH=$(ARCH) $(GOBUILD) -o build/$(BINARY_NAME)-linux-$(ARCH) . || echo "Failed to build for linux/$(ARCH)" ; \
	)
	tar -czvf $(BINARY_NAME)-linux-binaries.tar.gz -C build .

# Build for current platform
build-local:
	$(GOBUILD) -o $(BINARY_NAME) -v

# Build for all platforms
build-all: build-linux build-local

# Cross compilation
.PHONY: build test clean run deps build-linux build-local build-all