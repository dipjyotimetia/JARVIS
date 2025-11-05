VERSION = 0.0.3
# Optimized build flags for smaller binaries and better performance
LDFLAGS = -ldflags="-s -w -X 'github.com/dipjyotimetia/jarvis/cmd/cmd.Version=$(VERSION)'"
BUILDFLAGS = -trimpath
OUTDIR = ./dist

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

.PHONY: build
build:
	mkdir -p $(OUTDIR)
	rm -rf $(OUTDIR)/*
	go build $(BUILDFLAGS) -o $(OUTDIR) $(LDFLAGS) ./...

.PHONY: build-optimized
build-optimized:
	@echo "Building optimized binary (smaller size, no debug symbols)..."
	mkdir -p $(OUTDIR)
	rm -rf $(OUTDIR)/*
	CGO_ENABLED=0 go build $(BUILDFLAGS) -o $(OUTDIR)/jarvis $(LDFLAGS) .
	@echo "Binary size:"
	@ls -lh $(OUTDIR)/jarvis | awk '{print $$5 " " $$9}'

.PHONY: run
run: build
	./dist/jarvis generate-scenarios --path="specs/openapi/v3.0/mini_blog.yaml"

.PHONY: no-dirty
no-dirty:
	git diff --exit-code

# Performance and benchmarking targets
.PHONY: bench
bench:
	@echo "Running benchmarks for proxy and ollama components..."
	go test -bench=. -benchmem -run=^$$ ./internal/proxy/...
	go test -bench=. -benchmem -run=^$$ ./pkg/engine/ollama/...

.PHONY: bench-proxy
bench-proxy:
	@echo "Running proxy benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./internal/proxy/...

.PHONY: bench-ollama
bench-ollama:
	@echo "Running ollama engine benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./pkg/engine/ollama/...

.PHONY: profile
profile:
	@echo "Running CPU and memory profiling..."
	mkdir -p ./profiles
	go test -cpuprofile=./profiles/cpu.prof -memprofile=./profiles/mem.prof -bench=. ./internal/proxy/...
	@echo "Profiles saved to ./profiles/"
	@echo "To analyze: go tool pprof ./profiles/cpu.prof"
	@echo "To analyze: go tool pprof ./profiles/mem.prof"

.PHONY: profile-web
profile-web:
	@echo "Starting web-based profile analysis..."
	@if [ -f ./profiles/cpu.prof ]; then \
		echo "CPU Profile: http://localhost:8080"; \
		go tool pprof -http=:8080 ./profiles/cpu.prof; \
	else \
		echo "No CPU profile found. Run 'make profile' first."; \
	fi

.PHONY: test-performance
test-performance:
	@echo "Running performance regression tests..."
	go test -tags=performance -timeout=10m ./internal/proxy/...

.PHONY: bench-compare
bench-compare:
	@echo "Running benchmark comparison (requires benchstat)..."
	@if command -v benchstat >/dev/null 2>&1; then \
		echo "Running current benchmarks..."; \
		go test -bench=. -benchmem -count=5 ./internal/proxy/... > bench-new.txt; \
		echo "Compare with: benchstat bench-old.txt bench-new.txt"; \
	else \
		echo "Install benchstat: go install golang.org/x/perf/cmd/benchstat@latest"; \
	fi

.PHONY: install-bench-tools
install-bench-tools:
	@echo "Installing benchmark and profiling tools..."
	go install golang.org/x/perf/cmd/benchstat@latest
	go install github.com/google/pprof@latest
	@echo "Tools installed: benchstat, pprof"

# Testing targets
.PHONY: test
test:
	@echo "Running all tests..."
	go test -v -race -cover ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage report..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-short
test-short:
	@echo "Running short tests..."
	go test -short -v ./...

# Linting and code quality
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code is not formatted. Run 'make tidy' to fix."; \
		gofmt -l .; \
		exit 1; \
	fi

# Security scanning
.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/perf/cmd/benchstat@latest
	go install github.com/google/pprof@latest
	@echo "All tools installed!"

# CI target (runs all checks)
.PHONY: ci
ci: fmt-check lint test security
	@echo "All CI checks passed!"