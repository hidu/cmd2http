package main


import (
	"fmt"
	"log"
	 "net/http"
	 "flag"
	 "os"
	 "bytes"
	 "time"
	 "strings"
	 "regexp"
	 "encoding/json"
	 "os/exec"
	 "text/template"
	 jsonConf "github.com/daviddengcn/go-ljson-conf"
   )
var configPath=flag.String("conf","./cmd2http.conf","config file")

var port int

type param struct{
  name string
  defaultValue string
  isOption bool
}

type Conf struct{
   name string
   cmdStr string
   cmd string
   charset string
   params []*param
}

var confMap map[string]*Conf

var config *jsonConf.Conf

func main(){
   flag.Parse()
   logFile,_:=os.OpenFile("./cmd2http.log",os.O_CREATE|os.O_RDWR|os.O_APPEND,0666)
   defer logFile.Close()
   log.SetOutput(logFile)
//    log.SetFlags(log.LstdFlags|log.Lshortfile)
    loadConfig()
    
    startHttpServer()
}

func loadConfig(){
   log.Println("start load conf [",*configPath,"]")
   var err error
   config, err= jsonConf.Load(*configPath)
	if err != nil {
	  log.Println(err.Error(),config)
	  os.Exit(2)
	}
	port=config.Int("port",8310)
	
	confMap=make(map[string]*Conf)
	
	cmds:=config.Object("cmds",make(map[string]interface{}))
	
	for k,v:=range cmds{
	   _conf:=v.(map[string]interface{})
	   conf:=new(Conf)
	   conf.name=k
	   conf.charset="utf-8"
	   charset,has:=_conf["charset"]
	   if(has){
	      conf.charset=charset.(string)
	    }
	   conf.cmdStr,_=_conf["cmd"].(string)
	   conf.cmdStr=strings.TrimSpace(conf.cmdStr)
	   
	   ps:=regexp.MustCompile(`\s+`).Split(conf.cmdStr,-1)
	   conf.cmd=ps[0]
	   
	   for i:=1;i<len(ps);i++ {
	       item:=ps[i]
	        _param:=new(param)
	        _param.name=item
	       _param.isOption=true
	       
	       if(item[0]=='$'){
	        _param.isOption=false;
	        tmp:=strings.Split(item+"|","|")
	        _param.name=tmp[0][1:]
	        _param.defaultValue=tmp[1]
	       }
	       conf.params=append(conf.params,_param)
//	       fmt.Println(_param.name,_param.defaultValue)
	    }
	   log.Println("register[",k,"] cmd:",conf.cmdStr)
	   confMap[k]=conf
	}
	
	log.Println("load conf [",*configPath,"] finish [ok]")
}

func startHttpServer(){
//   http.ReadTimeout=60 * time.Second
   
   http.HandleFunc("/",myHandler_root)
   http.HandleFunc("/help",myHandler_help)
  
   addr:=fmt.Sprintf(":%d",port)
   log.Println("listen at",addr)
   fmt.Println("listen at",addr)
   
   http.ListenAndServe(addr,nil)
}

func Command(name string, args []string) *exec.Cmd {
	aname, err := exec.LookPath(name)
	if err != nil {
		aname = name
	}
	return &exec.Cmd{
		Path: aname,
		Args: args,
	}
}


func myHandler_root(w http.ResponseWriter, r *http.Request){
     startTime:=time.Now()
	  path:=strings.Trim(r.URL.Path,"/")
	  if(path==""){
	      myHandler_help(w,r)
	      return;
	   }
	  logStr:=r.RemoteAddr+" req:"+r.RequestURI
	  defer func(){
	       log.Println(logStr)
	   }()
	  
	  conf,has:=confMap[path]
	  if(!has) {
	     logStr=logStr+" not support cmd"
	     fmt.Fprintf(w,"not support")
	     return;
	  }
	  
	  args:=make([]string,len(conf.params))
	  for i,_param:=range conf.params{
		  if(_param.isOption){
	        args[i]=_param.name
		     continue
		  }
	     val:=r.FormValue(_param.name)
	     if(val==""){
	        val=_param.defaultValue
	      }
	      args[i]=val
	  }
	  cmd := Command(conf.cmd,args)
	  var out bytes.Buffer
		cmd.Stdout = &out
	  err := cmd.Run()
	  if err != nil {
	    log.Println(err)
	    fmt.Fprintf(w,err.Error())
	    return;
	  }
	  format:=r.FormValue("format")
	  str:=`<!DOCTYPE html><html><head>
	         <meta http-equiv='Content-Type' content='text/html; charset=%s' />
	         <title>%s cmd2http</title></head><body><pre>%s</pre></body></html>`
	         
	  outStr:=out.String()
	  logStr=logStr+fmt.Sprintf(" resLen:%d time_use:%v",len(outStr),time.Now().Sub(startTime))
	  
	  if(format=="" || format=="html"){
	    fmt.Fprintf(w,fmt.Sprintf(str,conf.charset,conf.name,outStr))
	   }else{
	    fmt.Fprintf(w,outStr)
	   }
}

func myHandler_help(w http.ResponseWriter, r *http.Request){
   str:=`<!DOCTYPE html><html>
         <head>
         <meta http-equiv='Content-Type' content='text/html; charset=utf-8' />
         <title>{{.title}} cmd2http</title>
          <script>
            function $(id){
              return document.getElementById(id);
               }
          function form_check(){
               var cmd=$('cmd').value;
               if(!cmd){
                    alert("pls choose cmd");
                    return false;
                    }
                var _param=$('cmd').value+"?"+$('params').value;
                $('div_url').innerHTML="http://"+location.host+"/"+_param;
                $('result').src=_param;
             }
          function cmd_change(){
                $('msg').innerHTML=msg[$('cmd').value]||"";
             }
          </script>
        </head><body>
          <h1>Help</h1>
          <div>
           <p>echo -n $wd $a $b|defaultValue </p>
          <p>
          http://localhost/<b>echo?wd=hello&a=world</b>
             ==&gt;   <b>#echo -n hello world defaultValue</b> 
          </p></div><br/>
          <form onsubmit='form_check();return false;'>
          cmd:<select id='cmd' onchange='cmd_change()'>
            <option value=''>pls choose cmd</option>
              {{.option_cmd}}
           </select>
               params:<input type='text' id='params' name='params' style="width:500px">
               <input type='submit'>
             <div id='msg'></div>
          </form>
          <script> var msg={{.msgs}}</script>
          <div id="div_url"></div>
          <iframe id='result' name="result" src="about:_blank" style="border:none;width:800px" onload="this.height=1500;" ></iframe>
          </body></html>`;
        
       msgs:=make(map[string]string)
       option_cmd:=""
       for name,_conf:=range confMap{
           option_cmd=option_cmd+"<option value="+name+">"+name+"</option>";
           msgs[name]=_conf.cmdStr
          }
        
	   title:=config.String("title","")
	   str=regexp.MustCompile(`\s+`).ReplaceAllString(str," ")
	   
	   tpl,_:=template.New("page").Parse(str)
	   values :=make(map[string]string)
	   values["title"]=title
	   values["option_cmd"]=option_cmd
	   
	   jsonByte,_:=json.Marshal(msgs)
	   values["msgs"]=string(jsonByte)
	   
	   tpl.Execute(w,values)
}
