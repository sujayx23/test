# Distributed Log Querying System - Server

This is a basic TCP server implementation for a distributed log querying system. The server listens for client connections, receives grep commands, executes them on local log files, and returns the results.

## Features

- TCP socket-based communication
- Configurable port and machine number
- Grep command execution using `popen()`
- Proper error handling for file not found and connection issues
- Results prefixed with machine identifier
- One client connection at a time

## Files

- `server.cpp` - Main server implementation
- `test_client.cpp` - Simple test client for verification
- `Makefile` - Build and test automation
- `machine.1.log` - Sample log file for testing

## Building

```bash
make all
```

Or build individually:
```bash
g++ -o server server.cpp
g++ -o test_client test_client.cpp
```

## Usage

### Server
```bash
./server <machine_num> <port>
```

Example:
```bash
./server 1 8080
```

This starts a server for machine 1 listening on port 8080. The server will look for a log file named `machine.1.log`.

### Test Client
```bash
./test_client <host> <port>
```

Example:
```bash
./test_client localhost 8080
```

## Testing

Run the automated test suite:
```bash
make test
```

This will:
1. Create sample log files
2. Start the server in background
3. Test various grep patterns
4. Clean up

## Protocol

1. Client connects to server
2. Client sends grep command (e.g., "-i error", "INFO", "WARN")
3. Server executes: `grep <command> machine.<num>.log`
4. Server sends results with "MACHINE_X:" prefix
5. Connection closes

## Example Session

```bash
# Terminal 1: Start server
./server 1 8080

# Terminal 2: Connect and query
echo "ERROR" | nc localhost 8080
# Output:
# MACHINE_1:2024-01-15 10:30:16 ERROR: Database connection failed
# MACHINE_1:2024-01-15 10:30:19 ERROR: File not found: config.xml
# MACHINE_1:2024-01-15 10:30:21 ERROR: Network timeout occurred
# MACHINE_1:2024-01-15 10:30:24 ERROR: Authentication failed
```

## Error Handling

- **File not found**: Returns error message with machine prefix
- **Invalid grep pattern**: Returns "No matches found" message
- **Connection failures**: Logs error and continues waiting for next client
- **Invalid arguments**: Shows usage information and exits

## Next Steps

This server is ready for integration into a distributed system where:
- Multiple servers run on different machines
- A client connects to all servers simultaneously
- Results are aggregated from all machines
- Fault tolerance is handled at the client level
