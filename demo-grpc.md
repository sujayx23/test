# gRPC Distributed Log Querying System - Demo

## System Overview

This Go-based gRPC system provides a modern, concurrent alternative to the C++ TCP implementation with the following improvements:

### Key Features

1. **Type-Safe Communication**: Uses Protocol Buffers for structured data exchange
2. **Concurrent Queries**: Client queries multiple servers simultaneously using goroutines
3. **Built-in Timeouts**: Prevents hanging on unresponsive servers
4. **Comprehensive Error Handling**: Graceful handling of server failures
5. **Security**: Input sanitization and command injection protection

## Architecture Comparison

### C++ TCP Version
```
Client ──TCP──► Server 1 ──► machine.1.log
Client ──TCP──► Server 2 ──► machine.2.log
Client ──TCP──► Server 3 ──► machine.3.log
(Sequential or manual threading)
```

### Go gRPC Version
```
Client ──gRPC──► Server 1 ──► machine.1.log
     ├──gRPC──► Server 2 ──► machine.2.log
     └──gRPC──► Server 3 ──► machine.3.log
(Concurrent goroutines)
```

## Installation and Setup

```bash
# 1. Install Go (if not already installed)
brew install go

# 2. Install protobuf compiler
brew install protobuf

# 3. Install Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 4. Build the system
make -f Makefile.grpc all
```

## Usage Examples

### Starting Servers
```bash
# Terminal 1: Start server for machine 1
./server-grpc -machine=1 -port=8080

# Terminal 2: Start server for machine 2  
./server-grpc -machine=2 -port=8081

# Terminal 3: Start server for machine 3
./server-grpc -machine=3 -port=8082
```

### Running Distributed Queries
```bash
# Basic error search across all servers
./client-grpc -pattern="ERROR" -servers="localhost:8080,localhost:8081,localhost:8082"

# Case-insensitive search
./client-grpc -pattern="error" -options="-i" -servers="localhost:8080,localhost:8081,localhost:8082"

# Regex pattern search
./client-grpc -pattern="[0-9]{4}-[0-9]{2}-[0-9]{2}" -options="-E" -servers="localhost:8080,localhost:8081,localhost:8082"

# Inverted search (exclude matches)
./client-grpc -pattern="DEBUG" -options="-v" -servers="localhost:8080,localhost:8081,localhost:8082"
```

## Expected Output

```
Querying 3 servers for pattern: 'ERROR'

=== Distributed Log Query Results ===
Pattern: ERROR
Servers queried: 3

✅ MACHINE_1: Found 4 matching lines in machine.1.log
   MACHINE_1:2024-01-15 10:30:16 ERROR: Database connection failed
   MACHINE_1:2024-01-15 10:30:19 ERROR: File not found: config.xml
   MACHINE_1:2024-01-15 10:30:21 ERROR: Network timeout occurred
   MACHINE_1:2024-01-15 10:30:24 ERROR: Authentication failed

✅ MACHINE_2: Found 2 matching lines in machine.2.log
   MACHINE_2:2024-01-15 10:31:16 ERROR: Queue processing failed
   MACHINE_2:2024-01-15 10:31:18 ERROR: Invalid credentials provided

✅ MACHINE_3: Found 2 matching lines in machine.3.log
   MACHINE_3:2024-01-15 10:32:16 ERROR: Rate limit exceeded
   MACHINE_3:2024-01-15 10:32:18 ERROR: Service unavailable

=== Summary ===
Total matching lines: 8
Successful servers: 3/3
Failed servers: 0
Total query time: 45ms
```

## Performance Benefits

### Concurrency
- **C++ TCP**: Sequential connections or manual threading
- **Go gRPC**: Automatic concurrent queries using goroutines

### Error Handling
- **C++ TCP**: Basic error checking
- **Go gRPC**: Comprehensive timeout and failure handling

### Type Safety
- **C++ TCP**: Manual string parsing and formatting
- **Go gRPC**: Generated protobuf code with compile-time type checking

### Scalability
- **C++ TCP**: Limited by manual connection management
- **Go gRPC**: Built-in connection pooling and efficient resource management

## Security Improvements

1. **Input Sanitization**: Removes dangerous characters from grep patterns
2. **Command Injection Protection**: Validates input before execution
3. **Timeout Protection**: Prevents long-running malicious patterns
4. **Regex Validation**: Ensures patterns are valid before execution

## Testing

```bash
# Run automated test suite
make -f Makefile.grpc test

# Manual testing
make -f Makefile.grpc run-server  # Start single server
make -f Makefile.grpc run-client  # Run client with defaults
```

## File Structure

```
g71_test/
├── logquery.proto          # gRPC service definition
├── server.go               # gRPC server implementation
├── client.go               # gRPC client implementation
├── go.mod                  # Go module dependencies
├── Makefile.grpc           # Build automation
├── README-grpc.md          # Documentation
├── install-go.sh           # Installation script
└── demo-grpc.md            # This demo file
```

## Next Steps

1. **Install Go and dependencies** using `./install-go.sh`
2. **Build the system** using `make -f Makefile.grpc all`
3. **Run tests** using `make -f Makefile.grpc test`
4. **Deploy multiple servers** and test distributed queries
5. **Scale up** to test with larger log files and more servers

This gRPC implementation provides a modern, scalable foundation for distributed log querying that's ready for production use.
