#include <iostream>
#include <string>
#include <cstring>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <cstdio>
#include <cstdlib>
#include <fstream>
#include <sstream>
#include <vector>
#include <algorithm>

class LogQueryServer {
private:
    int machine_num;
    int port;
    int server_socket;
    struct sockaddr_in server_addr;
    
public:
    LogQueryServer(int machine_num, int port) : machine_num(machine_num), port(port) {
        server_socket = -1;
    }
    
    ~LogQueryServer() {
        if (server_socket != -1) {
            close(server_socket);
        }
    }
    
    bool initialize() {
        // Create socket
        server_socket = socket(AF_INET, SOCK_STREAM, 0);
        if (server_socket < 0) {
            std::cerr << "Error: Failed to create socket" << std::endl;
            return false;
        }
        
        // Set socket options to allow reuse of address
        int opt = 1;
        if (setsockopt(server_socket, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
            std::cerr << "Error: Failed to set socket options" << std::endl;
            return false;
        }
        
        // Configure server address
        memset(&server_addr, 0, sizeof(server_addr));
        server_addr.sin_family = AF_INET;
        server_addr.sin_addr.s_addr = INADDR_ANY;
        server_addr.sin_port = htons(port);
        
        // Bind socket
        if (bind(server_socket, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
            std::cerr << "Error: Failed to bind to port " << port << std::endl;
            return false;
        }
        
        // Listen for connections
        if (listen(server_socket, 1) < 0) {
            std::cerr << "Error: Failed to listen on socket" << std::endl;
            return false;
        }
        
        std::cout << "Server started on machine " << machine_num << " listening on port " << port << std::endl;
        return true;
    }
    
    std::string executeGrep(const std::string& grep_command) {
        // Use absolute path to ensure we find the log file
        std::string log_file = "/Users/nithishsujay/g71_test/machine." + std::to_string(machine_num) + ".log";
        
        // Debug: Print current working directory and file path
        std::cout << "Looking for log file: " << log_file << std::endl;
        
        // Check if log file exists
        std::ifstream file(log_file);
        if (!file.good()) {
            return "MACHINE_" + std::to_string(machine_num) + ": Error: Log file '" + log_file + "' not found\n";
        }
        file.close();
        
        // Sanitize and validate grep command for security
        std::string sanitized_command = sanitizeGrepCommand(grep_command);
        if (sanitized_command.empty()) {
            return "MACHINE_" + std::to_string(machine_num) + ": Error: Invalid grep command\n";
        }
        
        // Construct the full grep command with filename output
        std::string full_command = "grep -H " + sanitized_command + " " + log_file;
        
        // Execute grep using popen
        FILE* pipe = popen(full_command.c_str(), "r");
        if (!pipe) {
            return "MACHINE_" + std::to_string(machine_num) + ": Error: Failed to execute grep command\n";
        }
        
        std::string result;
        char buffer[4096];  // Increased buffer size for better performance
        std::vector<std::string> matching_lines;
        int line_count = 0;
        
        // Read output from grep
        while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
            std::string line(buffer);
            // Remove trailing newline
            if (!line.empty() && line.back() == '\n') {
                line.pop_back();
            }
            if (!line.empty()) {
                matching_lines.push_back(line);
                line_count++;
            }
        }
        
        int status = pclose(pipe);
        
        // Format results with proper structure
        if (line_count > 0) {
            // Add header with line count
            result += "MACHINE_" + std::to_string(machine_num) + ": Found " + std::to_string(line_count) + " matching lines in " + log_file + "\n";
            
            // Add each matching line with machine prefix
            for (const auto& line : matching_lines) {
                result += "MACHINE_" + std::to_string(machine_num) + ":" + line + "\n";
            }
        } else {
            // No matches found
            result = "MACHINE_" + std::to_string(machine_num) + ": No matches found in " + log_file + " (0 lines)\n";
        }
        
        return result;
    }
    
    // Sanitize grep command to prevent command injection and ensure proper regex support
    std::string sanitizeGrepCommand(const std::string& command) {
        if (command.empty()) {
            return "";
        }
        
        // Remove any potential command injection attempts
        std::string sanitized = command;
        
        // Remove dangerous characters that could be used for command injection
        sanitized.erase(std::remove(sanitized.begin(), sanitized.end(), ';'), sanitized.end());
        sanitized.erase(std::remove(sanitized.begin(), sanitized.end(), '&'), sanitized.end());
        sanitized.erase(std::remove(sanitized.begin(), sanitized.end(), '|'), sanitized.end());
        sanitized.erase(std::remove(sanitized.begin(), sanitized.end(), '`'), sanitized.end());
        sanitized.erase(std::remove(sanitized.begin(), sanitized.end(), '$'), sanitized.end());
        sanitized.erase(std::remove(sanitized.begin(), sanitized.end(), '('), sanitized.end());
        sanitized.erase(std::remove(sanitized.begin(), sanitized.end(), ')'), sanitized.end());
        
        // Trim whitespace
        sanitized.erase(0, sanitized.find_first_not_of(" \t\r\n"));
        sanitized.erase(sanitized.find_last_not_of(" \t\r\n") + 1);
        
        return sanitized;
    }
    
    void run() {
        std::cout << "Waiting for client connection..." << std::endl;
        
        while (true) {
            // Accept client connection
            struct sockaddr_in client_addr;
            socklen_t client_len = sizeof(client_addr);
            
            int client_socket = accept(server_socket, (struct sockaddr*)&client_addr, &client_len);
            if (client_socket < 0) {
                std::cerr << "Error: Failed to accept client connection" << std::endl;
                continue;
            }
            
            std::cout << "Client connected from " << inet_ntoa(client_addr.sin_addr) 
                      << ":" << ntohs(client_addr.sin_port) << std::endl;
            
            // Read grep command from client with larger buffer
            char buffer[4096];
            memset(buffer, 0, sizeof(buffer));
            
            ssize_t bytes_received = recv(client_socket, buffer, sizeof(buffer) - 1, 0);
            if (bytes_received <= 0) {
                std::cerr << "Error: Failed to receive data from client" << std::endl;
                close(client_socket);
                continue;
            }
            
            std::string grep_command(buffer);
            // Remove trailing newline if present
            if (!grep_command.empty() && grep_command.back() == '\n') {
                grep_command.pop_back();
            }
            
            std::cout << "Received grep command: " << grep_command << std::endl;
            
            // Execute grep and get results
            std::string results = executeGrep(grep_command);
            
            // Send results back to client in chunks for large results
            const char* data = results.c_str();
            size_t total_bytes = results.length();
            size_t bytes_sent_total = 0;
            
            while (bytes_sent_total < total_bytes) {
                ssize_t bytes_sent = send(client_socket, data + bytes_sent_total, 
                                        total_bytes - bytes_sent_total, 0);
                if (bytes_sent < 0) {
                    std::cerr << "Error: Failed to send results to client" << std::endl;
                    break;
                }
                bytes_sent_total += bytes_sent;
            }
            
            if (bytes_sent_total == total_bytes) {
                std::cout << "Sent " << bytes_sent_total << " bytes to client" << std::endl;
            }
            
            // Close client connection
            close(client_socket);
            std::cout << "Client connection closed. Waiting for next connection..." << std::endl;
        }
    }
};

void printUsage(const char* program_name) {
    std::cout << "Usage: " << program_name << " <machine_num> <port>" << std::endl;
    std::cout << "Example: " << program_name << " 1 8080" << std::endl;
}

int main(int argc, char* argv[]) {
    if (argc != 3) {
        std::cerr << "Error: Invalid number of arguments" << std::endl;
        printUsage(argv[0]);
        return 1;
    }
    
    int machine_num = std::atoi(argv[1]);
    int port = std::atoi(argv[2]);
    
    if (machine_num <= 0 || port <= 0 || port > 65535) {
        std::cerr << "Error: Invalid machine number or port" << std::endl;
        printUsage(argv[0]);
        return 1;
    }
    
    LogQueryServer server(machine_num, port);
    
    if (!server.initialize()) {
        std::cerr << "Error: Failed to initialize server" << std::endl;
        return 1;
    }
    
    // Run the server
    server.run();
    
    return 0;
}
