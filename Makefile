GOBIN := $(shell /usr/bin/go env GOPATH)/bin
GARBLE := $(GOBIN)/garble
GARBLE_FLAGS := -literals -seed=random

# Source files
SOURCES := $(wildcard *.go)
BUILD_DIR := build

compressed-build: $(SOURCES)
	go build -x -ldflags='-s -w -extldflags "-static"' -o liveapi-runner .
	upx --ultra-brute liveapi-runner

garble-compressed-build: $(SOURCES)
	@if [ ! -f "$(GARBLE)" ]; then \
		echo "Installing Garble..."; \
		/usr/bin/go install mvdan.cc/garble@latest; \
	fi
	GOGARBLE=simple-no-memguard GOROOT=/usr/local/go $(GARBLE) $(GARBLE_FLAGS) build -x \
		-ldflags='-s -w' -o liveapi-runner .
	upx --ultra-brute liveapi-runner

garble-build: $(SOURCES)
	@if [ ! -f "$(GARBLE)" ]; then \
		echo "Installing Garble..."; \
		/usr/bin/go install mvdan.cc/garble@latest; \
	fi
	GOGARBLE=simple-no-memguard GOROOT=/usr/local/go $(GARBLE) $(GARBLE_FLAGS) build -x \
		-ldflags='-s -w' -o liveapi-runner .

build: $(SOURCES)
	go build -a -ldflags '-extldflags "-static"' -o liveapi-runner .

build-plugin:
	@echo "Building and encrypting plugin..."
	@cd build && go run build_plugin.go

build-all: build-plugin build

run: build
	./liveapi-runner

test-memory:
	@echo "Testing memory extraction vulnerability..."
	@./test_memory_extraction.sh

test-memory-simple:
	@echo "Testing memory extraction vulnerability (simple method)..."
	@./test_memory_simple.sh

monitor:
	@echo "Starting API monitoring..."
	@./monitor_api.sh

monitor-simple:
	@echo "Starting simple API monitoring..."
	@./simple_monitor.sh

clean:
	rm -f liveapi-runner