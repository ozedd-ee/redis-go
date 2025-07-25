APP_NAME=redis-go
PKG=./commands/
B_PKGS=./commands ./serializer
BINARY=redis-go
DOCKER_IMAGE=redis-go-server
DOCKER_CONTAINER=redis-go-container

.PHONY: all build run restart test bench cover lint fmt clean docker-build docker-run docker-stop

# Compile Go binary
build:
	go build -o $(BINARY)

# Run Go server locally
run:
	go run main.go

# Unit tests
test:
	go test $(PKG) -v

# Benchmarks
bench:
	go test -bench=. -benchmem -v $(B_PKGS)

# Remove binary
clean:
	rm -f $(BINARY)

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run:
	docker run --rm -d -p 6379:6379 --name $(DOCKER_CONTAINER) $(DOCKER_IMAGE)

# Stop Docker container
docker-stop:
	-docker stop $(DOCKER_CONTAINER)
