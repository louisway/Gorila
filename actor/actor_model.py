import torch
import torch.optim as optim
import numpy as np
import pickle
from model_component.nets import nets
from model_component.config import config
from py_com.client import socket_client
from py_com.server import socket_server
from model_component.replay import cartEnv


def server_get_request(svr):
    svr.accept()
    svr.getdata()
    request = svr.data.decode('ascii')
    print("Receiving Requeset from Go interface: " + request)
    return request

def select_action(model,state,ranseed):
    sample = random.random()
    if sample > ranseed:
        return model(Variable(FloatTensor([state]),volatile=True).type(FloatTensor)).data.max(1)[1].view(1,1)
    else:
        return LongTensor([[random.randrange(2)]])

def encode_data(state,action,next_state,reward):
    experience = []
    experience.append(state,tobytes())
    experience.append(action.tobytes())
    experience.append(next_state.tobytes())
    experience.append(reward.tobytes())
    return pickle.dumps(experience)
 
if __name__=='__main__':
    '''
    '''
    cfg = config()
    mdl = nets(cfg.Input_Dim, cfg.Hidden_Dim, cfg.Output_Dim)
    env = cartEnv()
    addr = ('localhost',10000)
    server = socket_server(addr)
    server.listen()
    go_addr = ('localhost',10001)
    client = socket_client(go_addr)
    output = client.connect()
    req = server_get_request(server)
    while req != "Model_Ready":
        req = server_get_request(server)
    mdl.load_state_dict(torch.load('./model/cartpole.mdl'))
    steps_done = 0
    while True:
        experience = []
        state = env.reset()
        eps_threshold = cfg.Eps_End + (cfg.Eps_Start-cfg.Eps_End)*math.exp(-1.*steps_done/cfg.Eps_Decay)
        while True:
            action = select_action(mdl,state,eps_threshold)
            action = action.item()
            action_np = np.array([action],dtype=np.int32)
            next_state,reward,done,_ = env.step(action)
            if done:
                reward = -1.0
            reward = np.array([reward],dtype=np.float64)
            b = encode_data(state,action_np,next_state,reward)
            experience.append(b)
            state = next_state
            steps_done = steps_done + 1
            if done:
                break; 
        client.send(bytes(str(len(experience)),'ascii')) 
        for e in experience:
            client.send(e)
        req = server_get_request(server)
        if req == "Update":
            mdl.load_state_dict(torch.load('./model/cartpole.mdl'))
