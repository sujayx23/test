# Distributed Log Querying System - gRPC Version

A Go-based distributed log querying system using gRPC for high-performance, concurrent log searching across multiple machines.

## Features

- **gRPC Communication**: Type-safe, efficient communication between client and servers
- **Concurrent Queries**: Client queries multiple servers simultaneously using goroutines
- **Fault Tolerance**: Handles server failures and timeouts gracefully
- **Full Grep Support**: Supports all grep options including regex patterns with `-E` flag
- **Line Counting**: Reports exact number of matching lines from each server
- **File Identification**: Each result includes the source filename
- **Command Injection Protection**: Sanitizes input patterns for security

## Architecture

```
┌─────────────┐    gRPC     ┌─────────────┐
│   Client    │ ──────────► │   Server 1  │ ──► machine.1.log
│             │             │             │
│             │    gRPC     ┌─────────────┐
│             │ ──────────► │   Server 2  │ ──► machine.2.log
│             │             │             │
│             │    gRPC     ┌─────────────┐
│             │ ──────────► │   Server 3  │ ──► machine.3.log
└─────────────┘             └─────────────┘
```

## Files

- `logquery.proto` - gRPC service definition
- `server.go` - gRPC server implementation
- `client.go` - gRPC client for distributed queries
- `Makefile.grpc` - Build and test automation
- `go.mod` - Go module dependencies

## Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
- Go protobuf plugins:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

## Building

```bash
# Install dependencies
make -f Makefile.grpc deps

# Build everything
make -f Makefile.grpc all

# Or build individually
make -f Makefile.grpc proto    # Generate protobuf code
make -f Makefile.grpc server   # Build server
make -f Makefile.grpc client   # Build client
```

## Usage

### Server

```bash
./server-grpc -machine=1 -port=8080
```

Options:
- `-machine`: Machine ID (used for log file naming: `machine.X.log`)
- `-port`: Port to listen on (default: 8080)

### Client

```bash
./client-grpc -pattern="ERROR" -servers="localhost:8080,localhost:8081,localhost:8082"
```

Options:
- `-pattern`: Grep pattern to search for (required)
- `-options`: Grep options (e.g., "-i", "-E", "-v")
- `-servers`: Comma-separated list of server addresses
- `-timeout`: Timeout for each server query (default: 10s)

## Examples

### Basic Text Search
```bash
./client-grpc -pattern="ERROR" -servers="localhost:8080,localhost:8081"
```

### Case-Insensitive Search
```bash
./client-grpc -pattern="error" -options="-i" -servers="localhost:8080,localhost:8081"
```

### Regex Pattern Search
```bash
./client-grpc -pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}" -options="-E" -servers="localhost:8080,localhost:8081"
```

### Inverted Search (exclude matches)
```bash
./client-grpc -pattern="DEBUG" -options="-v" -servers="localhost:8080,localhost:8081"
```

## Testing

Run the automated test suite:
```bash
make -f Makefile.grpc test
```

This will:
1. Build the server and client
2. Create sample log files for 3 machines
3. Start 3 servers on different ports
4. Run various query tests
5. Clean up

## gRPC Service Definition

```protobuf
service LogQuery {
    rpc QueryLogs(QueryRequest) returns (QueryResponse);
}

message QueryRequest {
    string pattern = 1;        // The grep pattern
    string options = 2;        // Grep options
    string machine_id = 3;     // Machine identifier
}

message QueryResponse {
    string machine_id = 1;     // Machine that processed query
    int32 line_count = 2;      // Number of matching lines
    string filename = 3;       // Log file name
    repeated string lines = 4; // Matching log lines
    string error = 5;          // Error message if any
    bool success = 6;          // Query success status
}
```

## Performance Features

- **Concurrent Server Queries**: All servers are queried simultaneously
- **Connection Pooling**: Efficient gRPC connection management
- **Timeout Handling**: Prevents hanging on unresponsive servers
- **Streaming Results**: Large result sets are handled efficiently
- **Memory Efficient**: Results are processed incrementally

## Error Handling

The system handles various error conditions:

- **Server Unavailable**: Reports connection failures
- **Timeout**: Configurable timeout per server query
- **Invalid Patterns**: Sanitizes and validates input
- **File Not Found**: Reports missing log files
- **Grep Errors**: Handles grep execution failures

## Security Features

- **Input Sanitization**: Removes dangerous characters from patterns
- **Command Injection Protection**: Validates and escapes input
- **Regex Validation**: Ensures patterns are valid before execution
- **Timeout Protection**: Prevents long-running malicious patterns

## Comparison with C++ Version

| Feature | C++ TCP | Go gRPC |
|---------|---------|---------|
| Communication | Raw TCP sockets | gRPC with protobuf |
| Concurrency | Single-threaded | Goroutines |
| Type Safety | Manual parsing | Generated code |
| Error Handling | Basic | Comprehensive |
| Timeout Support | None | Built-in |
| Performance | Good | Excellent |
| Maintainability | Moderate | High |

## Next Steps

This gRPC implementation provides a solid foundation for:

1. **Load Balancing**: Add load balancer support
2. **Authentication**: Implement TLS and authentication
3. **Caching**: Add result caching for repeated queries
4. **Monitoring**: Add metrics and health checks
5. **Scaling**: Support for thousands of servers
