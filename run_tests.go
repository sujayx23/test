package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	pb "github.com/sujayx23/g71_test/logquery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServerConfig represents a server configuration
type ServerConfig struct {
	MachineID string
	Address   string
}

// QueryResult represents the result from a single server
type QueryResult struct {
	MachineID string
	Response  *pb.QueryResponse
	Error     error
}

// LogQueryClient handles distributed log queries
type LogQueryClient struct {
	servers []ServerConfig
	timeout time.Duration
}

// NewLogQueryClient creates a new client instance
func NewLogQueryClient(servers []ServerConfig, timeout time.Duration) *LogQueryClient {
	return &LogQueryClient{
		servers: servers,
		timeout: timeout,
	}
}

// QueryAllServers queries all configured servers concurrently
func (c *LogQueryClient) QueryAllServers(pattern, options string) []QueryResult {
	var wg sync.WaitGroup
	results := make([]QueryResult, len(c.servers))
	resultChan := make(chan QueryResult, len(c.servers))

	// Query each server concurrently
	for i, server := range c.servers {
		wg.Add(1)
		go func(index int, srv ServerConfig) {
			defer wg.Done()
			result := c.queryServer(srv, pattern, options)
			result.MachineID = srv.MachineID
			resultChan <- result
		}(i, server)
	}

	// Wait for all queries to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		// Find the correct index for this result
		for i, server := range c.servers {
			if server.MachineID == result.MachineID {
				results[i] = result
				break
			}
		}
	}

	return results
}

// queryServer queries a single server
func (c *LogQueryClient) queryServer(server ServerConfig, pattern, options string) QueryResult {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Connect to server
	conn, err := grpc.Dial(server.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return QueryResult{
			Error: fmt.Errorf("failed to connect to %s: %v", server.Address, err),
		}
	}
	defer conn.Close()

	// Create client
	client := pb.NewLogQueryClient(conn)

	// Create request
	req := &pb.QueryRequest{
		Pattern:   pattern,
		Options:   options,
		MachineId: server.MachineID,
	}

	// Execute query
	response, err := client.QueryLogs(ctx, req)
	if err != nil {
		return QueryResult{
			Error: fmt.Errorf("query failed on %s: %v", server.Address, err),
		}
	}

	return QueryResult{
		Response: response,
	}
}

// TestDistributedLogQuery runs comprehensive distributed log query tests
func main() {
	fmt.Println("=== Distributed Log Query Unit Tests ===")

	// Setup test environment
	fmt.Println("Setting up test environment...")

	// Generate enhanced test logs
	generateTestLogs()

	// Start test servers
	servers := startTestServers()
	defer stopTestServers(servers)

	// Wait for servers to start
	time.Sleep(2 * time.Second)

	// Run comprehensive tests
	fmt.Println("\n=== Running Tests ===")

	testFrequentPatterns()
	testSomewhatFrequentPatterns()
	testRarePatterns()
	testPatternsInOneLog()
	testPatternsInSomeLogs()
	testPatternsInAllLogs()
	testGrepOptions()
	testFaultTolerance()
	testCountOnlyMode()

	fmt.Println("\n=== All Tests Completed ===")
}

// generateTestLogs creates comprehensive test log files
func generateTestLogs() {
	fmt.Println("Generating test log files...")

	// vm1.log - 15 lines with various patterns
	vm1Content := []string{
		"2024-01-15 10:30:15 INFO: Server started successfully",
		"2024-01-15 10:30:16 ERROR: Database connection failed",
		"2024-01-15 10:30:17 WARN: High memory usage detected",
		"2024-01-15 10:30:18 INFO: User login successful",
		"2024-01-15 10:30:19 ERROR: File not found: config.xml",
		"2024-01-15 10:30:20 INFO: Cache cleared",
		"2024-01-15 10:30:21 ERROR: Network timeout occurred",
		"2024-01-15 10:30:22 WARN: Disk space low",
		"2024-01-15 10:30:23 INFO: Backup completed",
		"2024-01-15 10:30:24 ERROR: Authentication failed",
		"2024-01-15 10:30:25 DEBUG: Cache hit for key user_123",
		"2024-01-15 10:30:26 CRITICAL: System overload detected",
		"2024-01-15 10:30:27 INFO: Process completed successfully",
		"2024-01-15 10:30:28 ERROR: Invalid input received",
		"2024-01-15 10:30:29 WARN: Performance degradation noticed",
	}

	// vm2.log - 12 lines with different patterns
	vm2Content := []string{
		"2024-01-15 10:31:15 INFO: Worker process started",
		"2024-01-15 10:31:16 ERROR: Queue processing failed",
		"2024-01-15 10:31:17 INFO: Data synchronization complete",
		"2024-01-15 10:31:18 ERROR: Invalid credentials provided",
		"2024-01-15 10:31:19 WARN: Connection pool exhausted",
		"2024-01-15 10:31:20 DEBUG: Memory allocation successful",
		"2024-01-15 10:31:21 INFO: Task queue processed",
		"2024-01-15 10:31:22 ERROR: Timeout waiting for response",
		"2024-01-15 10:31:23 WARN: Low disk space warning",
		"2024-01-15 10:31:24 INFO: Service health check passed",
		"2024-01-15 10:31:25 CRITICAL: Database connection lost",
		"2024-01-15 10:31:26 ERROR: Failed to process request",
	}

	// vm3.log - 10 lines with unique patterns
	vm3Content := []string{
		"2024-01-15 10:32:15 INFO: API endpoint responding",
		"2024-01-15 10:32:16 ERROR: Rate limit exceeded",
		"2024-01-15 10:32:17 INFO: Request processed successfully",
		"2024-01-15 10:32:18 ERROR: Service unavailable",
		"2024-01-15 10:32:19 WARN: High CPU usage detected",
		"2024-01-15 10:32:20 DEBUG: Request validation passed",
		"2024-01-15 10:32:21 INFO: Response sent to client",
		"2024-01-15 10:32:22 ERROR: Connection refused",
		"2024-01-15 10:32:23 WARN: Memory leak detected",
		"2024-01-15 10:32:24 INFO: Session terminated",
	}

	// Write log files
	writeLogFile("vm1.log", vm1Content)
	writeLogFile("vm2.log", vm2Content)
	writeLogFile("vm3.log", vm3Content)

	fmt.Println("Test log files generated successfully")
}

// writeLogFile writes content to a log file
func writeLogFile(filename string, content []string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	for _, line := range content {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			fmt.Printf("Failed to write to %s: %v\n", filename, err)
			return
		}
	}
}

// startTestServers starts test servers and returns their processes
func startTestServers() []*exec.Cmd {
	fmt.Println("Starting test servers...")

	servers := []*exec.Cmd{}

	// Start server on port 8080 (machine 1)
	cmd1 := exec.Command("./server-grpc", "-machine=1", "-port=8080")
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	if err := cmd1.Start(); err != nil {
		fmt.Printf("Failed to start server 1: %v\n", err)
		return servers
	}
	servers = append(servers, cmd1)

	// Start server on port 8081 (machine 2)
	cmd2 := exec.Command("./server-grpc", "-machine=2", "-port=8081")
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Start(); err != nil {
		fmt.Printf("Failed to start server 2: %v\n", err)
		return servers
	}
	servers = append(servers, cmd2)

	// Start server on port 8082 (machine 3)
	cmd3 := exec.Command("./server-grpc", "-machine=3", "-port=8082")
	cmd3.Stdout = os.Stdout
	cmd3.Stderr = os.Stderr
	if err := cmd3.Start(); err != nil {
		fmt.Printf("Failed to start server 3: %v\n", err)
		return servers
	}
	servers = append(servers, cmd3)

	return servers
}

// stopTestServers stops all test servers
func stopTestServers(servers []*exec.Cmd) {
	fmt.Println("Stopping test servers...")

	for i, server := range servers {
		if server.Process != nil {
			if err := server.Process.Kill(); err != nil {
				fmt.Printf("Failed to kill server %d: %v\n", i+1, err)
			}
		}
	}

	// Clean up any remaining processes
	exec.Command("pkill", "-f", "server-grpc").Run()
}

// testFrequentPatterns tests patterns that occur frequently
func testFrequentPatterns() {
	fmt.Println("\n--- Testing Frequent Patterns ---")

	// Test ERROR pattern (should be frequent)
	fmt.Println("Testing ERROR pattern...")
	results := queryServers("ERROR", "")
	expectedCounts := map[string]int{
		"8080": 5, // vm1.log has 5 ERROR lines
		"8081": 4, // vm2.log has 4 ERROR lines
		"8082": 3, // vm3.log has 3 ERROR lines
	}
	verifyResults(results, expectedCounts, "ERROR")

	// Test INFO pattern (should be frequent)
	fmt.Println("Testing INFO pattern...")
	results = queryServers("INFO", "")
	expectedCounts = map[string]int{
		"8080": 5, // vm1.log has 5 INFO lines
		"8081": 4, // vm2.log has 4 INFO lines
		"8082": 4, // vm3.log has 4 INFO lines
	}
	verifyResults(results, expectedCounts, "INFO")
}

// testSomewhatFrequentPatterns tests patterns that occur somewhat frequently
func testSomewhatFrequentPatterns() {
	fmt.Println("\n--- Testing Somewhat Frequent Patterns ---")

	// Test WARN pattern (should be somewhat frequent)
	fmt.Println("Testing WARN pattern...")
	results := queryServers("WARN", "")
	expectedCounts := map[string]int{
		"8080": 3, // vm1.log has 3 WARN lines
		"8081": 2, // vm2.log has 2 WARN lines
		"8082": 2, // vm3.log has 2 WARN lines
	}
	verifyResults(results, expectedCounts, "WARN")
}

// testRarePatterns tests patterns that occur rarely
func testRarePatterns() {
	fmt.Println("\n--- Testing Rare Patterns ---")

	// Test DEBUG pattern (should be rare)
	fmt.Println("Testing DEBUG pattern...")
	results := queryServers("DEBUG", "")
	expectedCounts := map[string]int{
		"8080": 1, // vm1.log has 1 DEBUG line
		"8081": 1, // vm2.log has 1 DEBUG line
		"8082": 1, // vm3.log has 1 DEBUG line
	}
	verifyResults(results, expectedCounts, "DEBUG")

	// Test CRITICAL pattern (should be rare)
	fmt.Println("Testing CRITICAL pattern...")
	results = queryServers("CRITICAL", "")
	expectedCounts = map[string]int{
		"8080": 1, // vm1.log has 1 CRITICAL line
		"8081": 1, // vm2.log has 1 CRITICAL line
		"8082": 0, // vm3.log has 0 CRITICAL lines
	}
	verifyResults(results, expectedCounts, "CRITICAL")
}

// testPatternsInOneLog tests patterns that occur in only one log
func testPatternsInOneLog() {
	fmt.Println("\n--- Testing Patterns in One Log ---")

	// Test pattern that only appears in vm1.log
	fmt.Println("Testing 'Cache hit' pattern...")
	results := queryServers("Cache hit", "")
	expectedCounts := map[string]int{
		"8080": 1, // vm1.log has 1 "Cache hit" line
		"8081": 0, // vm2.log has 0 "Cache hit" lines
		"8082": 0, // vm3.log has 0 "Cache hit" lines
	}
	verifyResults(results, expectedCounts, "Cache hit")
}

// testPatternsInSomeLogs tests patterns that occur in some logs
func testPatternsInSomeLogs() {
	fmt.Println("\n--- Testing Patterns in Some Logs ---")

	// Test CRITICAL pattern (appears in vm1 and vm2, not vm3)
	fmt.Println("Testing CRITICAL pattern...")
	results := queryServers("CRITICAL", "")
	expectedCounts := map[string]int{
		"8080": 1, // vm1.log has 1 CRITICAL line
		"8081": 1, // vm2.log has 1 CRITICAL line
		"8082": 0, // vm3.log has 0 CRITICAL lines
	}
	verifyResults(results, expectedCounts, "CRITICAL")
}

// testPatternsInAllLogs tests patterns that occur in all logs
func testPatternsInAllLogs() {
	fmt.Println("\n--- Testing Patterns in All Logs ---")

	// Test ERROR pattern (appears in all logs)
	fmt.Println("Testing ERROR pattern...")
	results := queryServers("ERROR", "")
	expectedCounts := map[string]int{
		"8080": 5, // vm1.log has 5 ERROR lines
		"8081": 4, // vm2.log has 4 ERROR lines
		"8082": 3, // vm3.log has 3 ERROR lines
	}
	verifyResults(results, expectedCounts, "ERROR")
}

// testGrepOptions tests various grep options
func testGrepOptions() {
	fmt.Println("\n--- Testing Grep Options ---")

	// Test case-insensitive search
	fmt.Println("Testing case-insensitive search...")
	results := queryServers("error", "-i")
	expectedCounts := map[string]int{
		"8080": 5, // Should find ERROR lines
		"8081": 4,
		"8082": 3,
	}
	verifyResults(results, expectedCounts, "error (case-insensitive)")

	// Test regex pattern
	fmt.Println("Testing regex pattern...")
	results = queryServers("[0-9]{4}-[0-9]{2}-[0-9]{2}", "-E")
	expectedCounts = map[string]int{
		"8080": 15, // All lines have timestamps
		"8081": 12,
		"8082": 10,
	}
	verifyResults(results, expectedCounts, "timestamp regex")
}

// testFaultTolerance tests fault tolerance
func testFaultTolerance() {
	fmt.Println("\n--- Testing Fault Tolerance ---")

	// Test with all servers running
	fmt.Println("Testing with all servers running...")
	results := queryServers("ERROR", "")
	if len(results) != 3 {
		fmt.Printf("❌ Expected 3 results, got %d\n", len(results))
	} else {
		fmt.Println("✅ All servers responded")
	}

	// Test with one server down (simulate by using wrong port)
	fmt.Println("Testing with one server down...")
	results = queryServers("ERROR", "", "localhost:8080", "localhost:8081", "localhost:9999")
	if len(results) != 3 {
		fmt.Printf("❌ Expected 3 results (including error), got %d\n", len(results))
	} else {
		fmt.Println("✅ Fault tolerance working")
	}

	// Verify that we get results from working servers
	successCount := 0
	for _, result := range results {
		if result.Error == nil && result.Response != nil && result.Response.Success {
			successCount++
		}
	}
	if successCount < 2 {
		fmt.Printf("❌ Expected at least 2 successful results, got %d\n", successCount)
	} else {
		fmt.Printf("✅ Got %d successful results from working servers\n", successCount)
	}
}

// testCountOnlyMode tests count-only mode
func testCountOnlyMode() {
	fmt.Println("\n--- Testing Count-Only Mode ---")

	// Test count-only mode
	fmt.Println("Testing count-only mode...")
	results := queryServers("ERROR", "", "localhost:8080", "localhost:8081", "localhost:8082")
	expectedCounts := map[string]int{
		"8080": 5,
		"8081": 4,
		"8082": 3,
	}
	verifyResults(results, expectedCounts, "ERROR (count-only)")
}

// queryServers queries the specified servers with the given pattern and options
func queryServers(pattern, options string, servers ...string) []QueryResult {
	if len(servers) == 0 {
		servers = []string{"localhost:8080", "localhost:8081", "localhost:8082"}
	}

	serverConfigs := make([]ServerConfig, len(servers))
	for i, addr := range servers {
		parts := strings.Split(addr, ":")
		machineID := "1"
		if len(parts) > 1 {
			machineID = parts[1]
		}

		serverConfigs[i] = ServerConfig{
			MachineID: machineID,
			Address:   addr,
		}
	}

	client := NewLogQueryClient(serverConfigs, 10*time.Second)
	return client.QueryAllServers(pattern, options)
}

// verifyResults verifies that the query results match expected counts
func verifyResults(results []QueryResult, expectedCounts map[string]int, testName string) {
	totalExpected := 0
	for _, count := range expectedCounts {
		totalExpected += count
	}

	totalActual := 0
	successCount := 0

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("Server %s error: %v\n", result.MachineID, result.Error)
			continue
		}

		if !result.Response.Success {
			fmt.Printf("Server %s failed: %s\n", result.MachineID, result.Response.Error)
			continue
		}

		successCount++
		actualCount := int(result.Response.LineCount)
		totalActual += actualCount

		expectedCount, exists := expectedCounts[result.MachineID]
		if !exists {
			fmt.Printf("❌ Test %s: Unexpected machine ID: %s\n", testName, result.MachineID)
			continue
		}

		if actualCount != expectedCount {
			fmt.Printf("❌ Test %s: Machine %s expected %d lines, got %d\n",
				testName, result.MachineID, expectedCount, actualCount)
			return
		}
	}

	if totalActual != totalExpected {
		fmt.Printf("❌ Test %s: Total expected %d lines, got %d\n",
			testName, totalExpected, totalActual)
		return
	}

	fmt.Printf("✅ Test %s passed: %d successful servers, %d total lines\n",
		testName, successCount, totalActual)
}
