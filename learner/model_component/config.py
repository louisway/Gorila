class config:
    def __init__(self):
      self.capacity = 10000
      self.t_update = 100
      self.use_cuda = False 
      self.Episodes = 200
      self.Eps_Start = 0.9
      self.Eps_End  = 0.05
      self.Eps_Decay = 200
      self.Gamma = 0.8
      self.LearningRate = 0.001
      self.Batch_Size = 64
      self.Input_Dim = 4
      self.Hidden_Dim = 256 
      self.Output_Dim = 2

