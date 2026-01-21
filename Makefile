COVERAGE_PROFILE := coverage.out
COVERAGE_REPORT := coverage.html

# Targets
.PHONY: test coverage clean mock document run help

# Default target
all: test coverage

# Run tests with coverage
test:
	go test ./... -coverprofile=$(COVERAGE_PROFILE)

# Generate coverage report
coverage: test
	go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_REPORT)
	@echo "Coverage report generated at $(COVERAGE_REPORT)"

# Clean coverage files
clean:
	rm -f $(COVERAGE_PROFILE) $(COVERAGE_REPORT)

# Run the server
run:
	go run cmd/main.go

# build project
build:
	go build main.go

# Help message
help:
	@echo "Usage:"
	@echo "  make test        Run tests with coverage"
	@echo "  make coverage    Generate HTML coverage report"
	@echo "  make clean       Remove coverage files"
	@echo "  make run         Run the server"
	@echo "  make help        Display this help message"