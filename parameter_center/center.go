package main

import (
        "log"
        "net"
        "net/http"
        "net/rpc"
        "io/ioutil"
        "bufio"
        "fmt"
        "strconv"
        "os"
        "time"
)

//define the data structure to store center infomation


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

type CenterData struct{
        Minibatch int
        steps     int
        update    int
  
        model_dir string
        grad_dir  string
        
        replay    Address
        learners  []Address
        actors    []Address

        gready    []bool
}

//hyperparameters

var Minibatch       = 128
var Update          = 100

var Minmum_learners = 1
var Minmum_actors   = 1 

var Max_replay    = 1000000

var centerdata = CenterData{
                 Minibatch,
                 1,
                 Update,
                 "./model/",
                 "./grad/",
                 Address{}, 
                 []Address{},
                 []Address{},
                 []bool{},
              }

type Center int

type Identity struct{
          Id int
}

type ReplyString struct{
          Reply string
}

func (c *Center) Connect(ad Address, reply *Identity) error {
  var r Identity
  switch t := ad.Ty; t{
  case Learners:
      r.Id = len(centerdata.learners)
      centerdata.learners = append(centerdata.learners, ad)
      centerdata.gready = append(centerdata.gready, false)
  case Actors:
      r.Id = len(centerdata.actors)
      centerdata.actors = append(centerdata.actors, ad) 
  case Replays:
      r.Id = 0
      centerdata.replay = ad
  }
  *reply = r 
  return nil
}

func (c *Center) Gradient(content Content , reply *ReplyString) error {
  var r ReplyString 
  r.Reply = "successed!"
  err := ioutil.WriteFile(centerdata.grad_dir+"grad_"+strconv.Itoa(content.id)+".gd", content.data, 0644) 
  if err != nil {
    r.Reply = "failed" 
  }
  centerdata.gready[content.id] = true
  reply = &r
  return nil
}

func reset_gready() {
  for idx,_ := range centerdata.gready {
    centerdata.gready[idx] = false
  } 
}

func check_gready_status() bool{
  for _,ele := range centerdata.gready {
    if ele == false {
      return false 
    } 
  }
  return true
}
 

func check_status() bool{
  if centerdata.replay.Ty == Unassigned {
    return false 
  }
  if len(centerdata.learners) < Minmum_learners {
    return false
  }
  if  len(centerdata.actors) < Minmum_actors {
    return false
  }
  return true
}

func UpdateReplay() bool{
    client, err := rpc.DialHTTP("tcp", centerdata.replay.Ip+":"+centerdata.replay.Port)
    if err != nil {
       log.Fatal("Init replay error: ", err) 
    }
    reply := "status"
    client.Call("Replay.UpdateMaxNum", Max_replay,&reply)
    return true
}

func UpdateReplayAddress(worksets []Address, nets string) bool{
  for _,ele := range worksets {
    client, err := rpc.DialHTTP("tcp", ele.Ip+":"+ele.Port)
    if err != nil {
      log.Fatal("update replay error: ", err) 
    }
    reply := "status"
    //nets: Learner.UpdateReplayAddr  Actor.UpdateReplayAddr
    client.Call(nets, centerdata.replay, &reply)
  } 
  return true
}

func UpdateQnet(worksets []Address, nets string, model *Content) bool {
  for _,ele := range worksets {
    client, err := rpc.DialHTTP("tcp", ele.Ip+":"+ele.Port)
    if err != nil {
       log.Fatal("Init nets error: ", err) 
    }
    //nets: Learner.UpdateQnet ,Actor.UpdateQnet, Learner.UpdateTarget
    reply := "status"
    client.Call(nets, *model, &reply) 
  }
  return true
}

func GoLearn() {
  portion := int(Minibatch/(len(centerdata.learners)))
  for _,ele := range centerdata.learners{
    client, err := rpc.DialHTTP("tcp", ele.Ip+":"+ele.Port)
    if err != nil {
        log.Fatal("Go learn error: ", err)
    }
    reply := "status"
    client.Call("Learner.GoLearn",portion,&reply)
  }

}

func load_model(model *Content) bool{
  //needs to implement
  file, err := os.Open(centerdata.model_dir+"cartpole.mdl")
  if err != nil {
    fmt.Println(err)  
    return false
  } 
  defer file.Close()
  fileinfo, fileinfoerr := file.Stat()
  if fileinfoerr != nil{
    fmt.Println(fileinfoerr)
    return false 
  }
  filesize := fileinfo.Size()
  buffer := make([]byte, filesize)
  bytesread,bytesreaderr := file.Read(buffer)
   
  if int(bytesread) != int(filesize) || bytesreaderr != nil {
    fmt.Println(bytesreaderr)
    return false
  }
  model.data = buffer
  return true
}

func NetListenAndServe(ls net.Listener){
  error := http.Serve(ls, nil)
  if error != nil {
       log.Fatal("Error serving: ", error) 
  }
  fmt.Println("closed")
}

func getFromPy(ls net.Listener) string{
  c,err := ls.Accept()
  if err != nil {
      fmt.Println(err)
  }
  netData,err := bufio.NewReader(c).ReadString('\n')
  return netData 
}

func main() {
    py_server, err:= net.Listen("tcp4",":10000")
    if err != nil {
        fmt.Println(err)
        return 
    } 
    defer py_server.Close()
    //waiting for model to be ready
    center := new(Center)
    err = rpc.Register(center)
    if err != nil {
       log.Fatal("Format of service Center isn't correct.", err)
    }
    go_server, err:= net.Listen("tcp",":1234")
    rpc.HandleHTTP()
    //Py_ready := ""
    //for Py_ready = getFromPy(py_server); Py_ready != "Model_Ready"; Py_ready = getFromPy(py_server) {
    //}
    fmt.Println("Model is ready") 
    //make sure all components is ready
    go NetListenAndServe(go_server)
    for ready := check_status(); ready != true; {
      fmt.Println("Waiting for each component to be ready")
       time.Sleep(10)
       fmt.Println(len(centerdata.learners))
      //NetListenAndServe(go_server)
    }
    //initialize Learners and actors replay
    init_model := Content{}
    load_model(&init_model)
    go UpdateQnet(centerdata.learners,"Learner.InitQnet",&init_model)
    go UpdateQnet(centerdata.learners,"Learner.InitTarget",&init_model)
    go UpdateQnet(centerdata.actors,"Actor.InitQnet",&init_model)
    //initialize Replay address for learners and actors
    go UpdateReplayAddress(centerdata.learners, "Learner.UpdateReplayAddr")
    go UpdateReplayAddress(centerdata.actors, "Actor.UpdateReplayAddr")
    //initialize replay
    go UpdateReplay()  
    for true {
      reset_gready()
      go GoLearn()
      for ready:=check_gready_status(); ready == false;{
        NetListenAndServe(go_server) 
      } 
      //ask model to update gradient and save new model
      //update
      load_model(&init_model)
      centerdata.steps = centerdata.steps + 1 
      if centerdata.steps%centerdata.update == 0{
        //send update Ttargetfirst
        UpdateQnet(centerdata.learners,"Learner.InitTarget",&init_model)
      }
      go UpdateQnet(centerdata.learners,"Learner.InitQnet",&init_model)
      go UpdateQnet(centerdata.actors,"Actor.InitQnet",&init_model)
      
    }
   py_server.Close()
   go_server.Close()      
     
}
