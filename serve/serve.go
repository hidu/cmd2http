package serve
import (
 "net/http"
 "os"
  "github.com/hidu/goutils"
  "github.com/hidu/goutils/cache"
   jsonConf "github.com/daviddengcn/go-ljson-conf"
   "fmt"
   "log"
)
type Cmd2HttpServe struct{
   logFile *os.File
   logPath string
   cacheDirPath string
   Port int
   Charset_list []string 
   Charset string 
   Timeout int
   CmdConfs map[string]*Conf
   Config *jsonConf.Conf
   Cache cache.Cache
}

var version string="1.3"

func (cmd2 *Cmd2HttpServe)Run(){
   cmd2.ParseConfig()
   
   http.Handle("/s/",http.FileServer(http.Dir("./")))
   http.HandleFunc("/res/",myHandler_res)
   http.HandleFunc("/",cmd2.myHandler_root)
   http.HandleFunc("/help",cmd2.myHandler_help)
   
   addr:=fmt.Sprintf(":%d",cmd2.Port)
   log.Println("listen at",addr)
   cmd2.setupLog()
   defer cmd2.logFile.Close()
   
   err:=http.ListenAndServe(addr,nil)
   if(err!=nil){
       fmt.Println(err.Error())
       log.Println(err.Error())
     }
}
func (cmd2 *Cmd2HttpServe)setupLog(){
     cmd2.logFile,_=os.OpenFile(cmd2.logPath,os.O_CREATE|os.O_RDWR|os.O_APPEND,0644)
     log.SetOutput(cmd2.logFile)
      
     goutils.SetInterval(func(){
     if(!goutils.File_exists(cmd2.logPath)){
           cmd2.logFile.Close()
           cmd2.logFile,_=os.OpenFile(cmd2.logPath,os.O_CREATE|os.O_RDWR|os.O_APPEND,0644)
           log.SetOutput(cmd2.logFile)
         }
     },30)
}

func (cmd2 *Cmd2HttpServe)setupCache(){
    if (len(cmd2.cacheDirPath)>5){
       cache.SetDefaultCacheHandler(cache.NewFileCache(cmd2.cacheDirPath))
    } 
}
