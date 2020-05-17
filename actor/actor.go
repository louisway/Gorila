package main

import (
        "fmt"
        "log"
        "net"
        "net/http"
        "net/rpc"
        "io/ioutil"
        "strconv"
        "bufio"
)

//define the data structure
type Type int
 
const (
    Unassigned Type = 0
    Learners Type = 1
    Actors  Type = 2
    Replays  Type = 3
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

type Identity struct{
         Id int
}

type Data struct{
         data []byte
}

type ActorData struct {
          id        int
          net_dir   string 
          data_dir  string 

          replay    Address
          center    Address
          self      Address 
        
          model_status bool 
}

var actordata = ActorData{
             0,
             "./model/",
             "./data/",
             Address{},
             Address{"192.168.1.14", "1234", 0},
             Address{"192.168.1.14", "1236", 2},
             false,
             }


func reset_status() {
  actordata.model_status = false
}

type Actor int

type ReplyString struct{
      Reply string
}

func (a *Actor)UpdateReplayAddr(addr Address, reply *ReplyString) error{
    var r ReplyString 
    r.Reply = "succeed!"
    actordata.replay = addr
    *reply = r
    return nil
}

func (a *Actor)UpdateQnet(model Content, reply *ReplyString) error {
    var r ReplyString
    r.Reply = "succeed!"
    err := ioutil.WriteFile(actordata.net_dir+"model.mdl", model.data, 0644)
    if err != nil {
       r.Reply = "failed"
    }
    *reply = r
    actordata.model_status = true
   return err
 }

func SendExperience(d Data){
   client, err := rpc.DialHTTP("tcp",actordata.replay.Ip+":"+actordata.replay.Port)
   if err != nil {
       log.Fatal("Send Experience error: ", err)
   }
   var reply ReplyString
   reply.Reply = "status"
   client.Call("Replay.StoreExperience", d, &reply) 
}

func RequestConnect() {
   client, err := rpc.DialHTTP("tcp",actordata.center.Ip+":"+actordata.center.Port)
    if err != nil {
          log.Fatal("Request Connect error: ", err)
    }
    var reply Identity
    client.Call("Center.Connect",actordata.self, &reply)
    fmt.Println(reply.Id)
    actordata.id = reply.Id
}

func NetListenAndServe(){
  listener, e := net.Listen("tcp",":1236")
  if e != nil {
       log.Fatal("Listen error:",e)
  }
  error := http.Serve(listener, nil)
  if error != nil {
       log.Fatal("Error serving: ", error)
  }

}

func getFromPyString(ls net.Listener) string{
  c,err := ls.Accept()
  if err != nil {
      fmt.Println(err) 
  }
  netData, err := bufio.NewReader(c).ReadString('\n')
  return netData
}

func getFromPyBytes(ls net.Listener) []byte{
    var content []byte
    c,err := ls.Accept()
    if err != nil {
        fmt.Println(err) 
    }
    reader := bufio.NewReader(c)
    netData, isPrefix, err := reader.ReadLine()
    content = append(content,netData...) 
    for netData,isPrefix,err=reader.ReadLine(); isPrefix == true; netData,isPrefix,err=reader.ReadLine() {
        content = append(content,netData...)  
    }
    return content
}

func sendToPyBytes(msg []byte) {
    c, err := net.Dial("tcp", ":10001")
    if err != nil {
        fmt.Println(err)
        return 
    }
    c.Write(msg)
}

func main() {
    py_server, err := net.Listen("tcp4",":10000")
    if err != nil {
        fmt.Println(err)
        return 
    }
    defer py_server.Close()
     
    actor := new (Actor)
    err = rpc.Register(actor)
    if err != nil {
      log.Fatal("Format of service Center isn't correct.", err)
     }
    rpc.HandleHTTP()
    reset_status()
    RequestConnect()
    go NetListenAndServe()
    for ;actordata.model_status == false; {
    }
    reset_status()
    sendToPyBytes([]byte("Model_Ready"))
    
    for true {
       req := getFromPyString(py_server) 
       lens,err := strconv.Atoi(req)
       if err != nil {
          fmt.Println(err) 
       }
       for l:=0; l < lens;l=l+1 {
         msg := getFromPyBytes(py_server)
         var d Data 
         d.data = msg
         SendExperience(d)
       }
       mesg := "No Update"
       if actordata.model_status == true{
         //update the model
         actordata.model_status = false 
         mesg = "Update"
       }
       sendToPyBytes([]byte(mesg))
    }
    
}
