package serve
import (
  "net/http"
  "fmt"
  "strings"
  "encoding/json"
  "time"
  "html"
  "mime"
  "path/filepath"
)

func result_send(w http.ResponseWriter, r *http.Request,conf *Conf,outStr string,logStr string){
	  format:=r.FormValue("format")
	  str:=`<!DOCTYPE html><html><head>
	         <meta http-equiv='Content-Type' content='text/html; charset=%s' />
	         <title>%s cmd2http</title></head><body><pre>%s</pre></body></html>`
	         
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
	       path:=strings.Trim(r.URL.Path,"/")
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
    res,err:=GetRes(path)
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
    w.Write(LoadRes(path))
}

func myHandler_res(w http.ResponseWriter, r *http.Request){
    response_res(w,r,r.URL.Path)
}