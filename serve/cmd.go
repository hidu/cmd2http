package serve
import (
  "os/exec"
  "syscall"
  "net/http"
  "bytes"
  "fmt"
  "log"
  "time"
   "github.com/hidu/goutils/cache"
)
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


func exec_cmd(w http.ResponseWriter, r *http.Request,conf *Conf,args []string,logStr string,cacheKey string){
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
		
		if out.Len()>0 && conf.cache_life>3{
		    cache.Set(cacheKey,out.Bytes(),conf.cache_life)
		}
		result_send(w,r,conf,out.String(),logStr)
}
