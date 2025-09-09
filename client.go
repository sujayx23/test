package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
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

// PrintResults formats and prints the query results
func (c *LogQueryClient) PrintResults(results []QueryResult, pattern string) {
	fmt.Printf("\n=== Distributed Log Query Results ===\n")
	fmt.Printf("Pattern: %s\n", pattern)
	fmt.Printf("Servers queried: %d\n\n", len(results))

	totalLines := 0
	successfulServers := 0

	// Sort results by machine ID for consistent output
	sort.Slice(results, func(i, j int) bool {
		return results[i].MachineID < results[j].MachineID
	})

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ MACHINE_%s: Error - %v\n", result.MachineID, result.Error)
			continue
		}

		if !result.Response.Success {
			fmt.Printf("❌ MACHINE_%s: %s\n", result.MachineID, result.Response.Error)
			continue
		}

		successfulServers++
		lineCount := result.Response.LineCount
		totalLines += int(lineCount)

		if lineCount == 0 {
			fmt.Printf("🔍 MACHINE_%s: No matches found in %s (0 lines)\n", 
				result.MachineID, result.Response.Filename)
		} else {
			fmt.Printf("✅ MACHINE_%s: Found %d matching lines in %s\n", 
				result.MachineID, lineCount, result.Response.Filename)
			
			// Print matching lines
			for _, line := range result.Response.Lines {
				fmt.Printf("   MACHINE_%s:%s\n", result.MachineID, line)
			}
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("=== Summary ===\n")
	fmt.Printf("Total matching lines: %d\n", totalLines)
	fmt.Printf("Successful servers: %d/%d\n", successfulServers, len(results))
	fmt.Printf("Failed servers: %d\n", len(results)-successfulServers)
}

func main() {
	// Parse command line flags
	pattern := flag.String("pattern", "", "Grep pattern to search for (required)")
	options := flag.String("options", "", "Grep options (e.g., '-i', '-E', '-v')")
	servers := flag.String("servers", "localhost:8080,localhost:8081,localhost:8082", 
		"Comma-separated list of server addresses")
	timeout := flag.Duration("timeout", 10*time.Second, "Timeout for each server query")
	flag.Parse()

	// Validate required arguments
	if *pattern == "" {
		log.Fatal("Pattern is required. Use -pattern flag.")
	}

	// Parse server list
	serverList := strings.Split(*servers, ",")
	serverConfigs := make([]ServerConfig, len(serverList))
	
	for i, addr := range serverList {
		addr = strings.TrimSpace(addr)
		// Extract machine ID from address (assume port number as machine ID for demo)
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

	// Create client
	client := NewLogQueryClient(serverConfigs, *timeout)

	// Execute distributed query
	fmt.Printf("Querying %d servers for pattern: '%s'\n", len(serverConfigs), *pattern)
	if *options != "" {
		fmt.Printf("Using grep options: %s\n", *options)
	}

	start := time.Now()
	results := client.QueryAllServers(*pattern, *options)
	duration := time.Since(start)

	// Print results
	client.PrintResults(results, *pattern)
	fmt.Printf("Total query time: %v\n", duration)
}
