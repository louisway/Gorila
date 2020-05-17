import torch.nn as nn
import torch.nn.functional as F

def layer_init(layer, w_scale=1.0):
    nn.init.orthogonal_(layer.weight.data)
    layer.weight.data.mul_(w_scale)
    nn.init.constant_(layer.bias.data, 0)
    return layer

class nets(nn.Module):
    def __init__(self,input_dim,hidden_dim,output_dim):
        super(nets, self).__init__()
        self.l1 = layer_init(nn.Linear(input_dim,hidden_dim))
        self.l2 = layer_init(nn.Linear(hidden_dim, output_dim))

    def forward(self,x):
        x = F.relu(self.l1(x))
        x = self.l2(x)
        return x

 
