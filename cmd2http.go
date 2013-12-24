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
	 "io/ioutil"
	 "regexp"
	 "encoding/json"
	 "os/exec"
	 "mime"
	 "html"
	 "syscall"
	 "path/filepath"
	 "text/template"
	 jsonConf "github.com/daviddengcn/go-ljson-conf"
	  "github.com/cookieo9/resources-go/v2/resources"
   )
var configPath=flag.String("conf","./cmd2http.conf","config file")
var _port=flag.Int("port",0,"http server port,overwrite the port in the config file")
var _help=flag.Bool("help",false,"show help")

var port int
var version string

type param struct{
  name string
  defaultValue string
  isValParam bool
  values []string
  html string
}

type Conf struct{
   name string
   cmdStr string
   cmd string
   charset string
   params []*param
   intro string
   timeout int
   charset_list []string
   group string
}

var confMap map[string]*Conf

var config *jsonConf.Conf

var charset_list []string

var charset_default string

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
     logPath:="./cmd2http.log"
	  logFile,_:=os.OpenFile(logPath,os.O_CREATE|os.O_RDWR|os.O_APPEND,0666)
	  defer logFile.Close()
	  log.SetOutput(logFile)
	  
      myTimer(5,func(){
      if(!isFileExists(logPath)){
           logFile.Close()
           logFile,_=os.OpenFile(logPath,os.O_CREATE|os.O_RDWR|os.O_APPEND,0666)
           log.SetOutput(logFile)
         }
     })
    startHttpServer()
}

func myTimer(sec int,call func()){
	ticker:= time.NewTicker(time.Duration(sec)*time.Second)
   go func () {
		   for{
		   select {
		    case <-ticker.C:
            	call()
		      }
    	    }
    }()
}

func startHttpServer(){
//   http.ReadTimeout=60 * time.Second
   
   http.Handle("/s/",http.FileServer(http.Dir("./")))
   http.HandleFunc("/res/",myHandler_res)
   http.HandleFunc("/",myHandler_root)
   http.HandleFunc("/help",myHandler_help)
   
   addr:=fmt.Sprintf(":%d",port)
   log.Println("listen at",addr)
   
   err:=http.ListenAndServe(addr,nil)
   if(err!=nil){
       log.Println(err.Error())
     }
}

func (p *param)ToString() string{
    return fmt.Sprintf("name:%s,default:%s,isValParam:%x",p.name,p.defaultValue,p.isValParam);
}

func in_array(item string,arr []string) bool{
  for _,a:=range arr{
     if(a==item){
        return true
       }
   }
 return false
}
func printHelp(){
       fmt.Println("useage:")
       flag.PrintDefaults()
       fmt.Println("\nconfig demo:\n")
       fmt.Println(string(loadRes("res/conf/cmd2http.conf")))
}
func loadConfig(){
   version=getVersion()

//   log.Println("start load conf [",*configPath,"]")
//   log.Println("use conf:",*configPath)
   
   pathAbs,_:=filepath.Abs(*configPath)
   
   _,_err:= os.Open(pathAbs)
   if _err != nil {
       log.Println("config file not exists!",*configPath)
       printHelp()
       os.Exit(1)
    }
   os.Chdir(filepath.Dir(*configPath))
   var err error
   config, err= jsonConf.Load(*configPath)
	if err != nil {
	  log.Println(err.Error(),config)
	  os.Exit(2)
	}
	port=config.Int("port",8310)
	
	charset_list=config.StringList("charset_list",[]string{})
	
	charset_default=config.String("charset","utf-8");
	
	if(!in_array(charset_default,charset_list)){
	   charset_list=append(charset_list,charset_default);   
	}
	
	timeout:=config.Int("timeout",30)
	if(timeout<1){
	  timeout=1
	}
	
	confMap=make(map[string]*Conf)
	
	cmds:=config.Object("cmds",make(map[string]interface{}))
	
	for k,_:=range cmds{
	   conf_path_pre:="cmds."+k+"."
	   
	   conf:=new(Conf)
	   conf.name=k
	   conf.timeout=timeout
	   conf.group=config.String(conf_path_pre+"group","default")
	   
	   conf.charset=config.String(conf_path_pre+"charset",charset_default)
      conf.intro=config.String(conf_path_pre+"intro","")
	   
	   conf.charset_list=config.StringList(conf_path_pre+"charset_list",charset_list)
	   
	    if(!in_array(conf.charset,conf.charset_list)){
			   conf.charset_list=append(conf.charset_list,conf.charset);   
		}
	   
	   conf.timeout=config.Int(conf_path_pre+"timeout",timeout)

	   conf.cmdStr=config.String(conf_path_pre+"cmd","")
	   
	   conf.cmdStr=strings.TrimSpace(conf.cmdStr)
	   conf.params=make([]*param,0,10)
	   
	   ps:=regexp.MustCompile(`\s+`).Split(conf.cmdStr,-1)
//	   fmt.Println(ps)
	   conf.cmd=ps[0]
	   
	   for i:=1;i<len(ps);i++ {
	       item:=ps[i]
//	       fmt.Println("i:",i,item)
	        _param:=new(param)
	        _param.name=item
	       
	       if(item[0]=='$'){
	        _param.isValParam=true;
	        tmp:=strings.Split(item+"|","|")
	        _param.name=tmp[0][1:]
	        _param.defaultValue=tmp[1]
	       _param.html=config.String(conf_path_pre+"params."+_param.name+".html","")
	       _param.values=config.StringList(conf_path_pre+"params."+_param.name+".values",[]string{})
	        }
	       conf.params=append(conf.params,_param)
//	       fmt.Println(_param.name,_param.defaultValue)
	    }
	   log.Println("register[",k,"] cmd:",conf.cmdStr)
	   confMap[k]=conf
	}
	
	log.Println("load conf [",*configPath,"] finish [ok]")
}

func loadRes(path string) []byte{
     res,err:=getRes(path)
     if(err!=nil){
        return []byte{}
      }
     r,_:=res.Open()
     bf,err:=ioutil.ReadAll(r)
     if(err!=nil){
        log.Println("read res[",path,"] failed",err.Error())
      }
     return bf
}

func getRes(path string)(resources.Resource,error){
    path=strings.TrimLeft(path,"/")
    res,err:=resources.Find(path)
    if(err!=nil){
      log.Println("load res[",path,"] failed",err.Error())
      return nil,err
     }
     return res,nil
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

func isFileExists(path string) bool{
  _,err := os.Stat( path )
  return err==nil
}

func myHandler_root(w http.ResponseWriter, r *http.Request){
     startTime:=time.Now()
	  path:=strings.Trim(r.URL.Path,"/")
	  if(path==""){
			if isFileExists("./s/index.html") {
		     http.Redirect(w,r,"/s/",302)
			  return;
			}
	      myHandler_help(w,r)
	      return;
	   }else if(path=="favicon.ico"){
	      response_res(w,r,"res/css/favicon.ico")
	      return
	   }
	   
	  logStr:=r.RemoteAddr+" req:"+r.RequestURI+" "
	  defer func(){
	       logStr+=fmt.Sprintf(" time_use:%v",time.Now().Sub(startTime))
	       log.Println(logStr)
	   }()
	  
	  conf,has:=confMap[path]
	  if(!has) {
	     logStr=logStr+"not support cmd"
	     w.WriteHeader(404)
	     fmt.Fprintf(w,"<h1>404</h1>")
	     return;
	  }
	  
	  args:=make([]string,len(conf.params)+1)
	  for i,_param:=range conf.params{
		  if(!_param.isValParam){
	        args[i+1]=_param.name
		     continue
		  }
	     val:=r.FormValue(_param.name)
	     if(val==""){
	        val=_param.defaultValue
	      }
	      args[i+1]=val
	  }
	  cmd := Command(conf.cmd,args)
  	  var out bytes.Buffer
		cmd.Stdout = &out
		
  	  var outErr bytes.Buffer
		cmd.Stderr = &outErr
	   err:=cmd.Start()
	 

	  if(err!=nil){
	     logStr+="Error:"+err.Error()
	     fmt.Fprintf(w,err.Error())
	     return;
	   }
   	done := make(chan error)
		go func() {
		    done <- cmd.Wait()
		}()
		
		cc := w.(http.CloseNotifier).CloseNotify()
		
		isResonseOk:=true
		
		killCmd:=func(msg string){
			  if err := cmd.Process.Kill(); err != nil {
	            log.Println("failed to kill: ", err)
	           }
	        logStr+="[killed:"+msg+"]"
//	        log.Println(logStr)
	        isResonseOk=false
		}
		
		select {
		    case <-cc:
		         killCmd("client close")
		    case <-time.After(time.Duration(conf.timeout) * time.Second):
               killCmd("timeout")
//               w.WriteHeader();
          case <-done:
		}
		if(isResonseOk){
   		cmd_status := cmd.ProcessState.Sys().(syscall.WaitStatus)
   		logStr+=fmt.Sprintf(" [status:%d]",cmd_status.ExitStatus())
		}
	  
		if(!isResonseOk || !cmd.ProcessState.Success()){
		    w.WriteHeader(500)
		    w.Write([]byte("<h1>Error 500</h1><pre>"))
		    w.Write([]byte(logStr))
		    w.Write([]byte("\n\nStdOut:\n"))
		    w.Write(out.Bytes())
		    w.Write([]byte("\nErrOut:\n"))
		    w.Write(outErr.Bytes())
		    w.Write([]byte("</pre>"))
		    
		    return;
		}
		
	  format:=r.FormValue("format")
	  str:=`<!DOCTYPE html><html><head>
	         <meta http-equiv='Content-Type' content='text/html; charset=%s' />
	         <title>%s cmd2http</title></head><body><pre>%s</pre></body></html>`
	         
	  outStr:=out.String()
	  logStr+=fmt.Sprintf("resLen:%d ",len(outStr))
	  
	  charset:=r.FormValue("charset")
	  if(charset==""){
	     charset=conf.charset
	  }
	  if(format=="" || format=="html"){
	       w.Header().Set("Content-Type","text/html;charset="+charset)
	       if(format==""){
   	       fmt.Fprintf(w,str,conf.charset,conf.name,html.EscapeString(outStr))
	       }else{
	          w.Write([]byte(outStr))
	          w.Write([]byte("<script>window.postMessage && window.parent.postMessage('"+conf.name+"_height_'+document.body.scrollHeight,'*')</script>"))
	         }
	   }else if(format=="jsonp"){
          w.Header().Set("Content-Type","text/javascript;charset="+charset)
	       cb:=r.FormValue("cb")
	       if(cb==""){
	           cb="jsonp_form_"+path
	        }
	       m:=make(map[string]string)
	       m["data"]=outStr
	       jsonByte,_:=json.Marshal(m)
	       fmt.Fprintf(w,fmt.Sprintf(`%s(%s)`,cb,string(jsonByte)))
	   }else{ 
       w.Header().Set("Content-Type","text/plain;charset="+charset)
	    w.Write([]byte(outStr))
	   }
}

func response_res(w http.ResponseWriter,r *http.Request,path string){
    res,err:=getRes(path)
    if(err!=nil){
        w.WriteHeader(404)
        return;
     }
    finfo,_:=res.Stat()
    modtime:=finfo.ModTime()
    if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
			h := w.Header()
		 	delete(h, "Content-Type")
   		delete(h, "Content-Length")
   		w.WriteHeader(http.StatusNotModified)
   		return
   		}
   mimeType:= mime.TypeByExtension(filepath.Ext(path))
   if(mimeType!=""){
       w.Header().Set("Content-Type",mimeType)
     }
    w.Header().Set("Last-Modified",modtime.UTC().Format(http.TimeFormat))
    w.Write(loadRes(path))
}

func myHandler_res(w http.ResponseWriter, r *http.Request){
    response_res(w,r,r.URL.Path)
}

func myHandler_help(w http.ResponseWriter, r *http.Request){
       str:=string(loadRes("res/tpl/help.html"));
       tabs_bd:="<div class='bd'>";
       
       groups:=make(map[string][]string)
       
       for name,_conf:=range confMap{
           if _,_has:=groups[_conf.group];!_has{
            groups[_conf.group]=[]string{}
               }
           groups[_conf.group]=append(groups[_conf.group],name)
            
           tabs_bd+="\n\n<div class='cmd_div' id='div_"+name+"' style='display:none'>\n<form action='/"+name+"' methor='get' onsubmit='return form_check(this,\""+name+"\")' id='form_"+name+"'>\n";
           tabs_bd+="<div class='note note-g'><div><b>uri</b> :&nbsp;/"+name+"</div>"+
           "<div><b>command</b> :&nbsp;[&nbsp;"+_conf.cmdStr+
           "&nbsp;]&nbsp;<b>timeout</b> :&nbsp;"+fmt.Sprintf("%d",_conf.timeout)+"s</div>"
           if(_conf.intro!=""){
             tabs_bd=tabs_bd+"<div><b>intro</b> :&nbsp;&nbsp;"+_conf.intro+"</div>"
              }
           tabs_bd=tabs_bd+"</div>";
           tabs_bd=tabs_bd+"<fieldset><ul class='ul-1'>"
              for _,_param:=range _conf.params{
                if(_param.isValParam && _param.name!="charset" && _param.name!="format"){
                   placeholder:=""
                   if(_param.defaultValue!=""){
                      placeholder="placeholder='"+_param.defaultValue+"'"
                         }
                   if(_param.html!=""){
                      placeholder+=" "+_param.html
                          }
                   tabs_bd+="<li>"+_param.name+":"
                   if(len(_param.values)==0){
                      tabs_bd+="<input class='r-text p_"+_param.name+"' type='text' name='"+_param.name+"' "+placeholder+">";
                   }else{
                      tabs_bd+="<select class='r-select p_"+_param.name+"' name='"+_param.name+"' "+placeholder+">"
                       for _,_v:=range _param.values{
                        tabs_bd+="<option value=\""+_v+"\">"+_v+"</option>"
                              }
                      tabs_bd+="</select>";
                         }
                   tabs_bd+="</li>\n"
                     }
                   }
           tabs_bd+="<li>format:<select name='format'><option value=''>default</option><option value='html'>html</option><option value='plain'>plain</option><option value='jsonp'>jsonp</option></select></li>\n";
           if(len(_conf.charset_list)>1 && _conf.charset!="null"){
	           tabs_bd+="<li>charset:<select name='charset'>"
	           for _,_charset:=range _conf.charset_list{
	                   _selected:="";
	                   if(_charset==_conf.charset){
			                   _selected="selected=selected";
			                  }
	               tabs_bd+="<option value='"+_charset+"' "+_selected+">"+_charset+"</option>"
	              }
	           tabs_bd+="</select></li>"
           }
           
           tabs_bd+=`</ul><div class='c'></div>
           <center><input type='submit' class='btn'><span style='margin-right:50px'>&nbsp;</span><input type='reset' class='btn' onclick='form_reset(this.form)' title='reset the form and abort the request'></center>
           </fieldset><br/>
            <div class='div_url'></div>
            <iframe id='ifr_`+_conf.name+`' src='about:_blank' style='border:none;width:99%;height:20px' onload='ifr_load(this)'></iframe>
            <div class='result'></div>
            </form>
            </div>`;
          }
        
      tabs_str:=tabs_bd+"</div></div>";
      if(isFileExists("./s/my.css")){
        tabs_str+="<link  type='text/css' rel='stylesheet' href='/s/my.css'>";
        }
      
      if(isFileExists("./s/my.js")){
        tabs_str+="<script src='/s/my.js'></script>";
        }
        
      content_menu:="<dl id='main_menu'>"
      for groupName,names:=range groups{
      content_menu+="<dt>"+groupName+"</dt>"
         for _,name:=range names{
           content_menu+="<dd><a href='#"+name+"' onclick=\"show_cmd('"+name+"')\">"+name+"</a></dd>";
            }
        }
      content_menu+="</dl>"
	   title:=config.String("title","")
	   
	   reg:=regexp.MustCompile(`\s+`)
	   tabs_str=reg.ReplaceAllString(tabs_str," ")
	   
	   str=reg.ReplaceAllString(str," ")
	   
	   tpl,_:=template.New("page").Parse(str)
	   values :=make(map[string]string)
	   values["version"]=version
	   values["title"]=title
	   values["content_body"]=tabs_str
	   values["content_menu"]=content_menu
	   values["intro"]=config.String("intro","")
	   
	   
	   w.Header().Add("c2h",version)
	   tpl.Execute(w,values)
}

func getVersion() string{
   return strings.TrimSpace(string(loadRes("res/version")));
}
