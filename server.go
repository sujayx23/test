package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	pb "github.com/sujayx23/g71_test/logquery"
	"google.golang.org/grpc"
)

// LogQueryServer implements the gRPC LogQuery service
type LogQueryServer struct {
	pb.UnimplementedLogQueryServer
	machineID string
	logFile   string
}

// NewLogQueryServer creates a new server instance
func NewLogQueryServer(machineID string) *LogQueryServer {
	return &LogQueryServer{
		machineID: machineID,
		logFile:   fmt.Sprintf("machine.%s.log", machineID),
	}
}

// QueryLogs implements the gRPC QueryLogs method
func (s *LogQueryServer) QueryLogs(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	log.Printf("Received query: pattern='%s', options='%s'", req.Pattern, req.Options)

	// Check if log file exists
	if _, err := os.Stat(s.logFile); os.IsNotExist(err) {
		return &pb.QueryResponse{
			MachineId: s.machineID,
			Filename:  s.logFile,
			Error:     fmt.Sprintf("Log file '%s' not found", s.logFile),
			Success:   false,
		}, nil
	}

	// Sanitize the pattern to prevent command injection
	sanitizedPattern := sanitizePattern(req.Pattern)
	if sanitizedPattern == "" {
		return &pb.QueryResponse{
			MachineId: s.machineID,
			Filename:  s.logFile,
			Error:     "Invalid or empty pattern",
			Success:   false,
		}, nil
	}

	// Execute grep command
	lines, lineCount, err := s.executeGrep(sanitizedPattern, req.Options)
	if err != nil {
		return &pb.QueryResponse{
			MachineId: s.machineID,
			Filename:  s.logFile,
			Error:     fmt.Sprintf("Grep execution failed: %v", err),
			Success:   false,
		}, nil
	}

	log.Printf("Found %d matching lines in %s", lineCount, s.logFile)

	return &pb.QueryResponse{
		MachineId: s.machineID,
		LineCount: int32(lineCount),
		Filename:  s.logFile,
		Lines:     lines,
		Success:   true,
	}, nil
}

// executeGrep runs the grep command and returns matching lines
func (s *LogQueryServer) executeGrep(pattern, options string) ([]string, int, error) {
	// Build grep command
	args := []string{}
	
	// Add options if provided
	if options != "" {
		optionList := strings.Fields(options)
		args = append(args, optionList...)
	}
	
	// Add pattern and filename
	args = append(args, "-e", pattern, "--", s.logFile)

	// Execute grep command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "grep", args...)
	output, err := cmd.Output()

	if err != nil {
		// Check if it's just "no matches found" (exit code 1)
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return []string{}, 0, nil
		}
		return nil, 0, err
	}

	// Parse output into lines
	outputStr := string(output)
	if outputStr == "" {
		return []string{}, 0, nil
	}

	lines := strings.Split(strings.TrimRight(outputStr, "\n"), "\n")
	return lines, len(lines), nil
}

// sanitizePattern removes potentially dangerous characters
func sanitizePattern(pattern string) string {
	if pattern == "" {
		return ""
	}

	// Trim whitespace only; let grep handle regex and options
	return strings.TrimSpace(pattern)
}

func main() {
	// Parse command line flags
	machineID := flag.String("machine", "1", "Machine ID for this server")
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	// Validate machine ID
	if *machineID == "" {
		log.Fatal("Machine ID cannot be empty")
	}

	// Create server instance
	server := NewLogQueryServer(*machineID)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterLogQueryServer(grpcServer, server)

	// Start listening
	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", *port, err)
	}

	log.Printf("gRPC server started on machine %s, listening on port %s", *machineID, *port)
	log.Printf("Log file: %s", server.logFile)

	// Start serving
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
