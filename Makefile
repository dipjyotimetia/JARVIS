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