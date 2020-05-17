import torch
import torch.optim as optim
import numpy as np
import pickle
from model_component.nets import nets
from model_component.config import config
from py_com.client import socket_client
from py_com.server import socket_server

def server_get_request(svr):
    svr.accept()
    svr.getdata()
    request = svr.data.decode('ascii')
    print("Receiving Requeset from Go interface: " + request)
    return request

def decode_data(data):
    experience = pickle.loads(data)
    state = np.frombuffer(experience[0],dtype=np.float64)
    action = np.frombuffer(experience[1],dtype = np.int32).tolist()[0]
    next_state = np.frombuffer(experience[2],dtype=np.float64)
    reward = np.frombuffer(experience[3],dtype = np.float64)
    return state,action,next_state,reward

if __name__ == '__main__':
    '''
    '''
    cfg = config()
    mdl = nets(cfg.Input_Dim, cfg.Hidden_Dim, cfg.Output_Dim)
    tmdl = nets(cfg.Input_Dim, cfg.Hidden_Dim, cfg.Output_Dim)
    addr = ('localhost',10001)
    server = socket_server(addr)
    server.listen()
    go_addr = ('localhost',10001)
    client = socket_client(go_addr)
    output = client.connect()
    while True:
        req = server_get_request(server)
        while req != "Model_Ready":
            req = server_get_request(server)     
        mdl.load_state_dict(torch.load('./model/cartpole.mdl'))
        mdl.eval()
        req = server_get_request(server)
        if req == "Update_Target__Yes":
            tmdl.load_state_dict(mdl.state_dict())
        else:
            print("no need to update target")
         
        req = server_get_request(server)
   
        
    
