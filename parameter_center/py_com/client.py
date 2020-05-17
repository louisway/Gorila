import socket
import sys

class socket_client:
    def __init__(self,address):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.server_address = address 

    def changeAddress(self, ip, port):
        self.server_address = (ip,port)

    def connect(self):
        self.sock.connect(self.server_address)
        
    def send(self,data):
        self.sock.sendall(data)
    
    def close(self):
        self.sock.close()
        
