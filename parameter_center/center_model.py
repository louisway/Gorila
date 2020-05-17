import torch
import torch.optim as optim
import os
from model_component.nets import nets
from model_component.config import config
from py_com.client import socket_client
from py_com.server import socket_server

def CountFile(path):
    return os.listdir(path)

def SaveGrad(model,path):
   params = list(model.parameters())
   grad = [pa.grad for pa in params]
   torch.save(grad,path)

def LoadGrad(model,path):
   files = CountFile(path)
   params = list(model.parameters())
   for f in files:
     grad = torch.load('./grad/'+f)
     for i in range(len(params)):
         if grad[i] == None:
             continue
         if params[i].grad == None:
             params[i].grad = grad[i]
         else:
             params[i].grad = params[i].grad + grad[i] 
   
def server_get_request(svr):
    svr.accept()
    svr.getdata()
    request = svr.data.decode('ascii')  
    print("Receiving Requeset from Go interface: " + request) 
    return request
 
if __name__ == '__main__':
    '''
        prepare the model
    '''
    cfg = config()
    mdl = nets(cfg.Input_Dim, cfg.Hidden_Dim, cfg.Output_Dim) 
    model_saved = CountFile('./model')  
    if len(model_saved) == 0:
        torch.save(mdl.state_dict(),'./model/cartpole.mdl')        
    else: 
        mdl.load_state_dict(torch.load('./model/cartpole.mdl'))
        mdl.eval()
    #LoadGrad(mdl,'./grad')
    op = optim.Adam(mdl.parameters(),cfg.LearningRate)
    
    addr = ('localhost', 10001)
    server = socket_server(addr)
    server.listen()
    go_addr = ('localhost', 10000)
    client = socket_client(go_addr)
    output = client.connect()
    '''
       Tell the go interface the model is ready
    '''     
    client.send(bytes("Model_Ready",'ascii'))
    #req = server_get_request(server) 
     
    client.close() 
    server.close()
