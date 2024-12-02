.PHONY: all clean proto build test

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Project parameters
BINARY_NAME=p2p-service-discover
PROTO_DIR=proto
PROTO_GO_DIR=pkg/proto/pb
MAIN_DIR=examples/simple

# Protoc parameters
PROTOC=protoc

all: proto build

# Install protoc dependencies
install-proto-deps:
	@echo "Installing protoc dependencies..."
	$(GOCMD) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCMD) install github.com/jibuji/go-stream-rpc/cmd/protoc-gen-stream-rpc@latest

# Generate protobuf files
proto: install-proto-deps
	@echo "Generating protobuf files..."
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--stream-rpc_out=. --stream-rpc_opt=paths=source_relative \
		internal/protocol/proto/peerlist.proto

# Build the project
build: proto
	@echo "Building..."
	cd $(MAIN_DIR) && $(GOBUILD) -o ../../bin/$(BINARY_NAME) -v

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f $(PROTO_GO_DIR)/*.pb.go

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: calculator
calculator:
	protoc --go_out=. --go_opt=paths=source_relative \
		--stream-rpc_out=. --stream-rpc_opt=paths=source_relative \
		examples/calculator/proto/calculator.proto
	cd examples/calculator && go build -o ../../bin/calculator