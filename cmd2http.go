package main


import (
    "fmt"
    "log"
     "flag"
     "os"
     jsonConf "github.com/daviddengcn/go-ljson-conf"
//      "github.com/cookieo9/resources-go/v2/resources"
     "./serve"
     "path/filepath"
   )
   
var configPath=flag.String("conf","./cmd2http.conf","config file")
var _port=flag.Int("port",0,"http server port,overwrite the port in the config file")
var _help=flag.Bool("help",false,"show help")

var port int

var config *jsonConf.Conf

var logPath string

var useCache bool

func main(){
    flag.Parse()
    if(*_help){
      printHelp()
      os.Exit(0)
     }
    loadConfig()
    
    if(*_port>0){
       port=*_port
     }
     
    server:=new(serve.Cmd2HttpServe)
    server.Config=config
    server.Port=port
    server.Run()
}


func loadConfig(){
   pathAbs,_:=filepath.Abs(*configPath)
   log.Println("use config:",pathAbs)
   f,_err:= os.Open(pathAbs)
   f.Close()
   if _err != nil {
       log.Println("config file not exists!",*configPath)
       printHelp()
       os.Exit(1)
    }
   conf_dir:=filepath.Dir(pathAbs)
   os.Chdir(conf_dir)
   log.Println("chdir ",conf_dir)
   var err error
   config, err= jsonConf.Load(pathAbs)
    if err != nil {
      log.Println(err.Error(),config)
      os.Exit(2)
    }
}

func printHelp(){
       fmt.Println("useage:")
       flag.PrintDefaults()
       fmt.Println("\nconfig demo:\n")
       fmt.Println(string(serve.LoadRes("res/conf/cmd2http.conf")))
}
