package serve
import(
  "time"
  "net/http"
  "strings"
  "fmt"
  "log"
)
func (cmd2 *Cmd2HttpServe)myHandler_root(w http.ResponseWriter, r *http.Request){
     startTime:=time.Now()
	  path:=strings.Trim(r.URL.Path,"/")
	  if(path==""){
			if IsFileExists("./s/index.html") {
		     http.Redirect(w,r,"/s/",302)
			  return;
			}
	      cmd2.myHandler_help(w,r)
	      return;
	   }else if(path=="favicon.ico"){
	      response_res(w,r,"res/css/favicon.ico")
	      return
	   }
	  logStr:=r.RemoteAddr+" req:"+r.RequestURI+" "
	  access_log:=func(){
	       logStr+=fmt.Sprintf(" time_use:%v",time.Now().Sub(startTime))
	       log.Println(logStr)
	   }
	  
	  conf,has:=cmd2.CmdConfs[path]
	  if(!has) {
	     logStr=logStr+"not support cmd"
	     w.WriteHeader(404)
	     fmt.Fprintf(w,"<h1>404</h1>")
	     access_log()
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
	  use_cache:=r.FormValue("cache")
	  cacheKey:=""
//	  fmt.Println("conf.cache_life",conf.cache_life)
	  if use_cache!="no" && conf.cache_life>3{
		  cacheKey=GetCacheKey(conf.cmd,args)
//		  log.Println("cache_key:",cacheKey)
		  cache_has,cache_data:=cmd2.Cache.Get(cacheKey)
		  if cache_has{
		     logStr+=" cache hit"
		     w.Header().Add("cache_hit","1")
	   	  result_send(w,r,conf,string(cache_data),logStr)
	   	  access_log()
	   	  return
		  }
	  }
	 exec_cmd(w,r,conf,args,logStr,cacheKey)
	 access_log()  
}