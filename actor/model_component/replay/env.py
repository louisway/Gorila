import gym
from gym import wrappers
from gym import wrappers

class cartEnv:
    def __init__(self):
        self.env = gym.make('CartPole-v0')

    def reset(self):
        return self.env.reset()
 
    def step(self,action):
      return self.env.step(action)
   
    def close(self):
      self.env.close()  
   
    def wrappers(self, directory,force=True):
      self.env = wrappers.Monitor(self.env, directory=directory,force=True)  

    def render(self, close = True, mode='rgb_array'):
      self.env.render(close=c,mode=mode)
