.PHONY: build test lint bench bench-stress clean run

# Build all packages
build:
	go build ./...

# Run all tests
test:
	go test ./... -count=1 -timeout 120s

# Run tests with verbose output
test-v:
	go test ./... -count=1 -timeout 120s -v

# Lint: vet + format check
lint:
	go vet ./...
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Files not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Run render engine benchmarks
bench:
	go test ./pkg/... -bench=. -benchmem -count=3 -run=^$$ -timeout 300s

# Run stress benchmarks only
bench-stress:
	go test ./pkg -bench=BenchmarkStress -benchmem -count=3 -run=^$$ -timeout 300s

# Run a Lua example (usage: make run FILE=examples/counter.lua)
FILE ?= examples/counter.lua
run:
	go run ./cmd/lumina $(FILE)

# Clean build artifacts
clean:
	go clean ./...

# Show documentation location
docs:
	@echo "Core API:      docs/API.md"
	@echo "Lux Components: docs/lux-api.md"
	@echo "Development:   docs/DEVELOPMENT.md"
	@echo "Testing:       docs/TESTING.md"
