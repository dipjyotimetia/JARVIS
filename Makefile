VERSION = 0.0.3
LDFLAGS = -ldflags="-X 'github.com/dipjyotimetia/jarvis/cmd/cmd.Version=$(VERSION)'"
OUTDIR = ./dist

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

.PHONY: build
build:
	mkdir -p $(OUTDIR)
	rm -rf $(OUTDIR)/*
	go build -o $(OUTDIR) $(LDFLAGS) ./...

.PHONY: run
run: build
	./dist/jarvis generate-scenarios --path="specs/openapi/v3.0/mini_blog.yaml"

.PHONY: no-dirty
no-dirty:
	git diff --exit-code

# Performance and benchmarking targets
.PHONY: bench
bench:
	@echo "Running benchmarks for proxy and pact components..."
	go test -bench=. -benchmem -run=^$$ ./internal/proxy/...
	go test -bench=. -benchmem -run=^$$ ./pkg/engine/pact/...
	go test -bench=. -benchmem -run=^$$ ./pkg/engine/ollama/...

.PHONY: bench-proxy
bench-proxy:
	@echo "Running proxy benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./internal/proxy/...

.PHONY: bench-pact
bench-pact:
	@echo "Running pact engine benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./pkg/engine/pact/...

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
	go test -tags=performance -timeout=10m ./pkg/engine/pact/...

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