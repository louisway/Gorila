package main

import (
         "log"
         "net"
         "net/http"
         "net/rpc"
)

//define the data structure

type Type int

const (
    Unassigned Type = 0
    Learners   Type = 1
    Actors     Type = 2
    Replays    Type = 3
)

type Address struct{
         Ip, Port string
         ty Type
}

type Content struct{
         data []byte
         ty   Type
         id   int
}

type Data struct{
          Dat [][]byte
}

type Request struct{
          size int          
}

type ReplyString struct{
          Reply string
}

type ReplayData struct {
          Max_Inventory int
          data          [][]byte
     
          learner       Address
          center        Address 
          self          Address
          
          ready         bool
}

var replaydata = ReplayData {
              0,
              [][]byte{},
              Address{},
              Address{"192.168.1.14","1234",1},
              Address{"192.168.1.14","1237",3},
              false, 
              false,
              }

type Replay int


func (r *Replay) UpdateMaxNum(Max_replay int, reply *ReplyString) error {
  var r ReplyString
  r.Reply = "succeed!"
  replaydata.Max_Inventory = Max_replay
  replaydata.ready = true
  reply = &r 
  return nil
}


func (r *Replay) StoreExperience(data Data, reply *ReplyString) error {
   var r ReplyString
   r.Reply = "succeed!"
   replaydata.data = append(replaydata.data,data.Dat)
   if len(replaydata.data) > Max_Inventory {
     replaydata.data = replaydata.data[len(replaydata.data)-Max_Inventory:]
   }
   reply = &r
   return nil
} 


func (r * Replay) RequestExperience(req Request, reply *Data) error {
   var r Data
   if len(replaydata.data) >= req.size {
       for i := 0; i < req.size; i = i+1{
           p := rand.Intn(len(replaydata.data))
           r.Dat = append(r.Dat,replaydata.data[p])   
       }  
   } 
   reply = &r
   return nil
}

func reset_status() {
   replaydata.request = false
}

func NetListenAndServe() {
  listener, e := net.Listen("tcp", ":1237")
  if e != nil {
          log.Fatal("Listen error:", e)
  }
  error := http.Serve(listener, nil)
  if error != nil {
       log.Fatal("Error serving: ", error)
  }

}


func RequestConnect() {
  client, err := rpc.DialHTTP("tcp",replaydata.center.Ip+":"+replaydata.center.Port)
   if err != nil {
         log.Fatal("Request Connect error: ", err)
     }
   reply := -1
   client.Call("Center.Connect",replaydata.self, &reply)
   replaydata.id = reply
}

func main() {
  replay := new(Replay)
  err := rpc.Register(learner)
  if err != nil {
     log.Fatal("Format of service Center isn't correct.", err)
    }
  rpc.HandleHTTP()
  RequestConnect()
  NetListenAndServe() 
}

