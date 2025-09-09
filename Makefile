CXX = g++
CXXFLAGS = -Wall -Wextra -std=c++11

# Default target
all: server test_client

# Build the server
server: server.cpp
	$(CXX) $(CXXFLAGS) -o server server.cpp

# Build the test client
test_client: test_client.cpp
	$(CXX) $(CXXFLAGS) -o test_client test_client.cpp

# Clean up compiled files
clean:
	rm -f server test_client

# Create sample log files for testing
sample_logs:
	@echo "Creating sample log files..."
	@echo "2024-01-15 10:30:15 INFO: Server started successfully" > machine.1.log
	@echo "2024-01-15 10:30:16 ERROR: Database connection failed" >> machine.1.log
	@echo "2024-01-15 10:30:17 WARN: High memory usage detected" >> machine.1.log
	@echo "2024-01-15 10:30:18 INFO: User login successful" >> machine.1.log
	@echo "2024-01-15 10:30:19 ERROR: File not found: config.xml" >> machine.1.log
	@echo "2024-01-15 10:30:20 INFO: Cache cleared" >> machine.1.log
	@echo "2024-01-15 10:30:21 ERROR: Network timeout occurred" >> machine.1.log
	@echo "2024-01-15 10:30:22 WARN: Disk space low" >> machine.1.log
	@echo "2024-01-15 10:30:23 INFO: Backup completed" >> machine.1.log
	@echo "2024-01-15 10:30:24 ERROR: Authentication failed" >> machine.1.log
	@echo "Sample log files created!"

# Test the server
test: server sample_logs
	@echo "Starting server in background..."
	@./server 1 8080 &
	@sleep 2
	@echo "Testing with 'ERROR' pattern..."
	@echo "ERROR" | nc localhost 8080
	@echo "Testing with 'INFO' pattern..."
	@echo "INFO" | nc localhost 8080
	@echo "Testing with 'nonexistent' pattern..."
	@echo "nonexistent" | nc localhost 8080
	@pkill -f "./server 1 8080" || true
	@echo "Test completed!"

.PHONY: all clean sample_logs test
