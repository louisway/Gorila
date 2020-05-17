package main

import (
        "fmt"
        "log"
        "net"
        "net/http"
        "net/rpc"
        "io/ioutil"
)

//define the data structure 

type Identity struct{
    Id int
}

type Type int

const (
   Unassigned Type = 0
   Learners   Type = 1
   Actors     Type = 2
   Replays    Type = 3
)

type Address struct{
        Ip, Port string
        Ty Type
}

type Content struct{
        data []byte
        ty   Type
        id   int
}

type Data struct{
        data [][]byte
}

type Request struct{
        size  int
}

type ReplyString struct{
        Reply string
}

type LearnerData struct {
         batch_size  int
         id          int
         
         net_dir     string
         Tnet_dir    string
         grad_dir    string
         data_dir    string
         
         replay      Address
         center      Address  
         self        Address  
      
         replay_status bool
         model_status  bool     
         Tmodel_status bool
         ready_to_learn bool
         data_status    bool
}

//hyperparameter



var learnerdata = LearnerData{
               0,
               0,
               "./model/",
               "./Tmodel/",
               "./grad/",
               "./data/",
               Address{}, //replay
               Address{"localhost","1234",Unassigned},//center
               Address{"localhost","1235",Learners},//self
               false,
               false,
               false,
               false,
               false,
              }

func reset_status() {
  //learnerdata.replay_status = false
  learnerdata.model_status    = false
  learnerdata.Tmodel_status  = false
  learnerdata.ready_to_learn = false 
  learnerdata.data_status     = false 
}

type Learner int

func (l *Learner)UpdateReplayAddr(addr Address, reply *ReplyString) error{
  var r ReplyString 
  r.Reply = "succeed!"
  learnerdata.replay = addr
  *reply = r 
  return nil
}

func (l *Learner)UpdateQnet(model Content, reply *ReplyString) error {
   var r ReplyString 
   r.Reply = "succeed!"
   err := ioutil.WriteFile(learnerdata.net_dir+"cartpole.mdl", model.data, 0644)
   if err != nil {
      r.Reply = "failed"
   }
   learnerdata.model_status = true 
   reply = &r
   return err
}

func (l *Learner)UpdateTarget(model Content, reply *ReplyString) error {
   var r ReplyString
   r.Reply = "succeed!"
   err := ioutil.WriteFile(learnerdata.Tnet_dir+"cartpole.mdl", model.data, 0644)
   if err != nil {
      r.Reply = "failed"
   }
   reply = &r
   learnerdata.Tmodel_status = true 
   return err
}

func (l *Learner)GetExperience(data Data, reply *ReplyString) error {
  var r ReplyString
  r.Reply = "succeed!"
  //need to update
  err := ioutil.WriteFile(learnerdata.data_dir+"data", data.data[0], 0644)
  if err != nil {
      r.Reply = "failed"
   }
   reply = &r
  learnerdata.data_status = true
  return err
}

func (l *Learner)GoLearn(batch int, reply *ReplyString) error{
  var r ReplyString
  r.Reply = "succeed!" 
  learnerdata.batch_size = batch
  learnerdata.ready_to_learn = true  
  *reply = r
  return nil
}

func load_grad(model *Content) bool {
  //load grad file
  return true
}

func NetListenAndServe() {
  listener, e := net.Listen("tcp", ":1235")
  if e != nil {
          log.Fatal("Listen error:", e)
  }
  error := http.Serve(listener, nil)
  if error != nil {
       log.Fatal("Error serving: ", error)
  }

}

func RequestExperience() {
  req := Request{learnerdata.self,learnerdata.batch_size}
  client, err := rpc.DialHTTP("tcp",learnerdata.replay.Ip+":"+learnerdata.replay.Port)
  if err != nil {
        log.Fatal("Request Experience error: ", err)
    }
    var reply  ReplyString
    reply.Reply = "status"
    client.Call("Replay.RequestExperience",req ,&reply)
}

func RequestConnect() {
  client, err := rpc.DialHTTP("tcp",learnerdata.center.Ip+":"+learnerdata.center.Port)
  if err != nil {
         log.Fatal("Request Connect error: ", err)
     }
   var reply Identity 
   client.Call("Center.Connect",learnerdata.self, &reply)
   fmt.Println(learnerdata.self.Ty)
   fmt.Println("Got an id from center: ")
   fmt.Println(reply.Id)
   learnerdata.id = reply.Id
}

func SendGrad() {
  grad := Content{ []byte{}, 1,learnerdata.id}
  //load grad from dir 
  client, err := rpc.DialHTTP("tcp",learnerdata.center.Ip+":"+learnerdata.center.Port)
  if err != nil {
         log.Fatal("Request Experience error: ", err)
     }
  var reply ReplyString
  reply.Reply = "status" 
  client.Call("Center.Gradient",grad, &reply)

}

func main() {
    learner := new(Learner)
    err := rpc.Register(learner)
    if err != nil {
     log.Fatal("Format of service Center isn't correct.", err)
    }
    rpc.HandleHTTP()
    RequestConnect()
    //if learnerdata.replay_status is not ready do something
    for true  {
       reset_status()
       for learnerdata.model_status == false {
         NetListenAndServe() 
       } 
       //if learnerdata.Tmode_status == true do something means the target network need to be updated
       for learnerdata.ready_to_learn == false {
         NetListenAndServe() 
       } 
       RequestExperience()
       for learnerdata.data_status == false {
         NetListenAndServe()
       }
       //train
       SendGrad() 
    }     

}
