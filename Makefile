.PHONY: build test clean run docker docker-run cross help

BINARY_NAME=tldexpand
DOCKER_IMAGE=tldexpand:latest

help:
	@echo "TLDexpand - Makefile targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run tests"
	@echo "  bench       - Run benchmarks"
	@echo "  run         - Build and run with example"
	@echo "  clean       - Remove build artifacts"
	@echo "  docker      - Build Docker image"
	@echo "  docker-run  - Run in Docker"
	@echo "  cross       - Cross-compile for multiple platforms"

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Done! Binary: ./$(BINARY_NAME)"

test:
	@echo "Running tests..."
	go test -v -cover

bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem

run: build
	@echo "Running example scan for 'google' with ccTLDs..."
	./$(BINARY_NAME) -d google -i ccTLDs.txt

clean:
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)
	rm -f tldexpand-*
	@echo "Done!"

docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Done! Image: $(DOCKER_IMAGE)"

docker-run:
	@echo "Running in Docker (google with ccTLDs)..."
	docker run --rm $(DOCKER_IMAGE) -d google -i ccTLDs.txt

cross:
	@echo "Cross-compiling for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe .
	@echo "Done! Built binaries:"
	@ls -lh $(BINARY_NAME)-*
