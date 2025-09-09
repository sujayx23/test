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
        std::string log_file = "machine." + std::to_string(machine_num) + ".log";
        
        // Check if log file exists
        std::ifstream file(log_file);
        if (!file.good()) {
            return "MACHINE_" + std::to_string(machine_num) + ": Error: Log file '" + log_file + "' not found\n";
        }
        file.close();
        
        // Construct the full grep command
        std::string full_command = "grep " + grep_command + " " + log_file;
        
        // Execute grep using popen
        FILE* pipe = popen(full_command.c_str(), "r");
        if (!pipe) {
            return "MACHINE_" + std::to_string(machine_num) + ": Error: Failed to execute grep command\n";
        }
        
        std::string result;
        char buffer[1024];
        
        // Read output from grep
        while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
            std::string line(buffer);
            // Remove trailing newline and add machine prefix
            if (!line.empty() && line.back() == '\n') {
                line.pop_back();
            }
            result += "MACHINE_" + std::to_string(machine_num) + ":" + line + "\n";
        }
        
        int status = pclose(pipe);
        if (status != 0 && result.empty()) {
            // No matches found or error occurred
            result = "MACHINE_" + std::to_string(machine_num) + ": No matches found\n";
        }
        
        return result;
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
            
            // Read grep command from client
            char buffer[1024];
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
            
            // Send results back to client
            ssize_t bytes_sent = send(client_socket, results.c_str(), results.length(), 0);
            if (bytes_sent < 0) {
                std::cerr << "Error: Failed to send results to client" << std::endl;
            } else {
                std::cout << "Sent " << bytes_sent << " bytes to client" << std::endl;
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
