#include <iostream>
#include <string>
#include <cstring>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <unistd.h>

int main(int argc, char* argv[]) {
    if (argc != 3) {
        std::cerr << "Usage: " << argv[0] << " <host> <port>" << std::endl;
        return 1;
    }
    
    const char* host = argv[1];
    int port = std::atoi(argv[2]);
    
    // Create socket
    int client_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (client_socket < 0) {
        std::cerr << "Error: Failed to create socket" << std::endl;
        return 1;
    }
    
    // Configure server address
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(port);
    
    if (inet_pton(AF_INET, host, &server_addr.sin_addr) <= 0) {
        // Try to resolve hostname if direct IP conversion fails
        struct hostent* he = gethostbyname(host);
        if (he == nullptr) {
            std::cerr << "Error: Invalid address or hostname" << std::endl;
            close(client_socket);
            return 1;
        }
        memcpy(&server_addr.sin_addr, he->h_addr_list[0], he->h_length);
    }
    
    // Connect to server
    if (connect(client_socket, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        std::cerr << "Error: Failed to connect to server" << std::endl;
        close(client_socket);
        return 1;
    }
    
    std::cout << "Connected to server. Enter grep command: ";
    std::string grep_command;
    std::getline(std::cin, grep_command);
    
    // Send grep command
    ssize_t bytes_sent = send(client_socket, grep_command.c_str(), grep_command.length(), 0);
    if (bytes_sent < 0) {
        std::cerr << "Error: Failed to send data" << std::endl;
        close(client_socket);
        return 1;
    }
    
    // Receive results
    char buffer[4096];
    std::string results;
    
    while (true) {
        ssize_t bytes_received = recv(client_socket, buffer, sizeof(buffer) - 1, 0);
        if (bytes_received <= 0) {
            break;
        }
        buffer[bytes_received] = '\0';
        results += buffer;
    }
    
    std::cout << "\nResults from server:\n" << results << std::endl;
    
    close(client_socket);
    return 0;
}
