#!/usr/bin/env python3
import socket
import sys

def test_server():
    try:
        # Create socket
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(5)
        
        # Connect to server
        sock.connect(('localhost', 8080))
        print("Connected to server")
        
        # Send command
        command = "ERROR\n"
        sock.send(command.encode())
        print(f"Sent: {command.strip()}")
        
        # Receive response
        response = sock.recv(4096).decode()
        print(f"Received: {response}")
        
        sock.close()
        
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_server()
