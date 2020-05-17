import socket
import sys

class socket_server:
    def __init__(self,address):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.address = address 
        self.sock.bind(self.address)
        self.data = bytearray()

    def listen(self):
        self.sock.listen(1)

    def accept(self):
        self.connection, self.client_address = self.sock.accept()
        print(self.client_address)

    def getdata(self):
        while True:
            d = self.connection.recv(16)
            if d:
                self.data.extend(d)
            else:
                break 
        return self.data

    def senddata(self,data):
        self.connection.sendall(data)

    def close(self):
        self.sock.close()   
