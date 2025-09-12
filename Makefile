# Makefile for Go gRPC Distributed Log Querying System

# Go module and protobuf setup
GO_MODULE = distributed-log-querying
PROTO_FILE = logquery.proto
PROTO_DIR = logquery

# Build targets
SERVER_BINARY = server-grpc
CLIENT_BINARY = client-grpc

# Default target
.PHONY: all
all: proto server client

# Generate protobuf Go code
.PHONY: proto
proto:
	@echo "Generating protobuf Go code..."
	@mkdir -p $(PROTO_DIR)
	@protoc --go_out=./$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=./$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)
	@echo "Protobuf code generated in $(PROTO_DIR)/"

# Build the gRPC server
.PHONY: server
server: proto
	@echo "Building gRPC server..."
	@go build -o $(SERVER_BINARY) server.go
	@echo "Server built: $(SERVER_BINARY)"

# Build the gRPC client
.PHONY: client
client: proto
	@echo "Building gRPC client..."
	@go build -o $(CLIENT_BINARY) client.go
	@echo "Client built: $(CLIENT_BINARY)"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing Go dependencies..."
	@go mod tidy
	@go mod download

# Create sample log files for testing
.PHONY: sample-logs
sample-logs:
	@echo "Creating sample log files..."
	@echo "2024-01-15 10:30:15 INFO: Server started successfully" > vm1.log
	@echo "2024-01-15 10:30:16 ERROR: Database connection failed" >> vm1.log
	@echo "2024-01-15 10:30:17 WARN: High memory usage detected" >> vm1.log
	@echo "2024-01-15 10:30:18 INFO: User login successful" >> vm1.log
	@echo "2024-01-15 10:30:19 ERROR: File not found: config.xml" >> vm1.log
	@echo "2024-01-15 10:30:20 INFO: Cache cleared" >> vm1.log
	@echo "2024-01-15 10:30:21 ERROR: Network timeout occurred" >> vm1.log
	@echo "2024-01-15 10:30:22 WARN: Disk space low" >> vm1.log
	@echo "2024-01-15 10:30:23 INFO: Backup completed" >> vm1.log
	@echo "2024-01-15 10:30:24 ERROR: Authentication failed" >> vm1.log
	
	@echo "2024-01-15 10:31:15 INFO: Worker process started" > vm2.log
	@echo "2024-01-15 10:31:16 ERROR: Queue processing failed" >> vm2.log
	@echo "2024-01-15 10:31:17 INFO: Data synchronization complete" >> vm2.log
	@echo "2024-01-15 10:31:18 ERROR: Invalid credentials provided" >> vm2.log
	@echo "2024-01-15 10:31:19 WARN: Connection pool exhausted" >> vm2.log
	
	@echo "2024-01-15 10:32:15 INFO: API endpoint responding" > vm3.log
	@echo "2024-01-15 10:32:16 ERROR: Rate limit exceeded" >> vm3.log
	@echo "2024-01-15 10:32:17 INFO: Request processed successfully" >> vm3.log
	@echo "2024-01-15 10:32:18 ERROR: Service unavailable" >> vm3.log
	@echo "Sample log files created!"

# Test the gRPC system
.PHONY: test
test: server client sample-logs
	@echo "Testing gRPC distributed log querying system..."
	@echo "Starting servers in background..."
	@./$(SERVER_BINARY) -machine=1 -port=8080 &
	@./$(SERVER_BINARY) -machine=2 -port=8081 &
	@./$(SERVER_BINARY) -machine=3 -port=8082 &
	@sleep 3
	@echo "Testing ERROR pattern..."
	@./$(CLIENT_BINARY) "ERROR" -servers="localhost:8080,localhost:8081,localhost:8082"
	@echo "Testing INFO pattern..."
	@./$(CLIENT_BINARY) "INFO" -servers="localhost:8080,localhost:8081,localhost:8082"
	@echo "Testing with grep options..."
	@./$(CLIENT_BINARY) "error" -options="-i" -servers="localhost:8080,localhost:8081,localhost:8082"
	@echo "Testing regex pattern..."
	@./$(CLIENT_BINARY) "[0-9]{4}-[0-9]{2}-[0-9]{2}" -options="-E" -servers="localhost:8080,localhost:8081,localhost:8082"
	@echo "Testing count-only mode..."
	@./$(CLIENT_BINARY) -c "ERROR" -servers="localhost:8080,localhost:8081,localhost:8082"
	@pkill -f $(SERVER_BINARY) || true
	@echo "Test completed!"

# Run Go unit tests
.PHONY: test-unit
test-unit: server client
	@echo "Running Go unit tests..."
	@go run run_tests.go
	@echo "Unit tests completed!"

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f $(SERVER_BINARY) $(CLIENT_BINARY)
	@rm -rf $(PROTO_DIR)
	@rm -f vm*.log
	@pkill -f $(SERVER_BINARY) || true
	@echo "Cleanup complete!"

# Run a single server for manual testing
.PHONY: run-server
run-server: server
	@echo "Starting gRPC server (machine 1, port 8080)..."
	@./$(SERVER_BINARY) -machine=1 -port=8080

# Run client with default settings
.PHONY: run-client
run-client: client
	@echo "Running gRPC client..."
	@./$(CLIENT_BINARY) "ERROR" -servers="localhost:8080"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all         - Build everything (proto, server, client)"
	@echo "  proto       - Generate protobuf Go code"
	@echo "  server      - Build gRPC server"
	@echo "  client      - Build gRPC client"
	@echo "  deps        - Install Go dependencies"
	@echo "  sample-logs - Create sample log files"
	@echo "  test        - Run automated tests"
	@echo "  test-unit   - Run Go unit tests"
	@echo "  run-server  - Start a single server"
	@echo "  run-client  - Run client with default settings"
	@echo "  clean       - Clean up build artifacts"
	@echo "  help        - Show this help"
