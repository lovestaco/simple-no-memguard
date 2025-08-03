GOBIN := $(shell /usr/bin/go env GOPATH)/bin
GARBLE := $(GOBIN)/garble
GARBLE_FLAGS := -literals -seed=random

compressed-build:
	CGO_ENABLED=0 go build -x -ldflags='-s -w -extldflags "-static"' -o liveapi-runner .
	upx --ultra-brute liveapi-runner

garble-compressed-build:
	@if [ ! -f "$(GARBLE)" ]; then \
		echo "Installing Garble..."; \
		/usr/bin/go install mvdan.cc/garble@latest; \
	fi
	GOGARBLE=runner CGO_ENABLED=0 GOROOT=/usr/local/go $(GARBLE) $(GARBLE_FLAGS) build -x \
		-ldflags='-s -w' -o liveapi-runner .
	upx --ultra-brute liveapi-runner

build:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o liveapi-runner .

run:
	./liveapi-runner

test-memory:
	@echo "Testing memory extraction vulnerability..."
	@./test_memory_extraction.sh

test-memory-simple:
	@echo "Testing memory extraction vulnerability (simple method)..."
	@./test_memory_simple.sh