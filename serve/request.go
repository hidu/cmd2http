package serve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Request struct {
	writer    http.ResponseWriter
	req       *http.Request
	req_path  string
	cmd2      *Cmd2HttpServe
	startTime time.Time
	logInfo   []string
	stop      bool
	cacheKey  string
	cmdConf   *cmdItem
	cmdArgs   []string
	cmdEnv    map[string]string
}

func (req *Request) Deal() {
	req.startTime = time.Now()
	req.logInfo = make([]string, 0, 20)
	req.req_path = strings.Trim(req.req.URL.Path, "/")

	//    fmt.Println("path:",req.req_path)

	req.log(req.req.RemoteAddr, req.req.RequestURI)
	defer req.Close()
	req.handleStatic()
	if req.stop {
		return
	}
	req.tryExecCmd()
}

func (req *Request) log(infos ...string) {
	for _, _info := range infos {
		_info = strings.TrimSpace(_info)
		if _info != "" {
			req.logInfo = append(req.logInfo, _info)
		}
	}
}

func (req *Request) Close() {
	used := fmt.Sprintf("time_use:%.3fms", float64(time.Now().Sub(req.startTime).Nanoseconds())/1000000.0)
	log.Println(strings.Join(req.logInfo, " "), used)
	//    fmt.Println(len(req.logInfo),req.logInfo)
}

func (req *Request) handleStatic() {
	if req.req_path == "" {
		if IsFileExists("./s/index.html") {
			http.Redirect(req.writer, req.req, "/s/", 302)
		} else {
			req.cmd2.myHandler_help(req.writer, req.req)
		}
		req.stop = true
	}
}

func (req *Request) tryExecCmd() {
	conf, has := req.cmd2.config.Cmds[req.req_path]
	if !has {
		req.log("status:404")
		req.writer.WriteHeader(404)
		fmt.Fprintf(req.writer, "<h1>404</h1>")
		return
	}

	args := make([]string, len(conf.paramsAll))
	env := make(map[string]string)

	_param_prefix := "c2h_form_"

	for i, _param := range conf.paramsAll {
		if !_param.isValParam {
			args[i] = _param.Name
			continue
		}
		val := req.req.FormValue(_param.Name)
		if val == "" {
			val = _param.DefaultValue
		}
		args[i] = val
		env[_param_prefix+_param.Name] = val
	}
	for k, v := range req.req.Form {
		_key := _param_prefix + k
		_, has := env[_key]
		if !has {
			env[_key] = strings.Join(v, "\t")
		}
	}
	req.cmdConf = conf
	req.cmdArgs = args
	req.cmdEnv = env

	use_cache := req.req.FormValue("cache")
	//      fmt.Println("conf.cache_life",conf.cache_life)
	if use_cache != "no" && conf.CacheLife > 3 {
		req.cacheKey = GetCacheKey(conf.CmdRaw, args)
		//          log.Println("cache_key:",cacheKey)
		cache_has, cache_data := req.cmd2.Cache.Get(req.cacheKey)
		if cache_has {
			req.log("cache hit")
			req.writer.Header().Add("cache_hit", "1")
			req.sendResponse(string(cache_data))
			return
		}
	}
	//      req.log("["+conf.cmd+" "+strings.Join(args," ")+"]")
	//      fmt.Println("args:",args)
	req.exec()
}

func (req *Request) exec() {
	conf := req.cmdConf
	cmd := exec.Command(conf.Cmd, req.cmdArgs...)
	//when use cache,disable the env params
	env := syscall.Environ()
	if conf.CacheLife < 3 {
		for k, v := range req.cmdEnv {
			env = append(env, k+"="+v)
		}
	}
	cmd.Env = env
	var out bytes.Buffer
	cmd.Stdout = &out

	var outErr bytes.Buffer
	cmd.Stderr = &outErr
	err := cmd.Start()

	if err != nil {
		req.log("error:" + err.Error())
		req.writer.WriteHeader(500)
		fmt.Fprintf(req.writer, err.Error()+"\n")
		for arg_i, arg_v := range req.cmdArgs {
			fmt.Fprintf(req.writer, "arg_%d:%s\n", arg_i, arg_v)
		}
		return
	}
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	cc := req.writer.(http.CloseNotifier).CloseNotify()

	isResonseOk := true

	killCmd := func(msg string) {
		if err := cmd.Process.Kill(); err != nil {
			log.Println("failed to kill: ", err)
		}
		req.log("killed:" + msg)
		//            log.Println(logStr)
		isResonseOk = false
	}

	select {
	case <-cc:
		killCmd("client close")
	case <-time.After(time.Duration(conf.Timeout) * time.Second):
		killCmd("timeout")
		//               w.WriteHeader();
	case <-done:
	}
	if isResonseOk {
		cmd_status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exit_status := fmt.Sprintf(" [status:%d]", cmd_status.ExitStatus())
		req.log(exit_status)
	}

	if !isResonseOk || !cmd.ProcessState.Success() {
		req.writer.WriteHeader(500)
		req.writer.Write([]byte("<h1>Error 500</h1><pre>"))
		req.writer.Write([]byte(strings.Join(req.logInfo, " ")))
		req.writer.Write([]byte("\n\nStdOut:\n"))
		req.writer.Write(out.Bytes())
		req.writer.Write([]byte("\nErrOut:\n"))
		req.writer.Write(outErr.Bytes())
		req.writer.Write([]byte("</pre>"))
		return
	}

	if out.Len() > 0 && conf.CacheLife > 3 {
		req.cmd2.Cache.Set(req.cacheKey, out.Bytes(), conf.CacheLife)
	}
	req.sendResponse(out.String())
}

var tpl_html string = `<!DOCTYPE html><html><head>
         <meta http-equiv='Content-Type' content='text/html; charset=%s' />
          <title>%s cmd2http</title></head><body><pre>%s</pre></body></html>`

func (req *Request) sendResponse(outStr string) {
	format := req.req.FormValue("format")

	req.log(fmt.Sprintf("resLen:%d ", len(outStr)))
	w := req.writer
	r := req.req
	conf := req.cmdConf
	//      fmt.Println("outStr:",outStr)
	charset := r.FormValue("charset")
	if charset == "" {
		charset = conf.Charset
	}
	if format == "" || format == "html" {
		w.Header().Set("Content-Type", "text/html;charset="+charset)
		if format == "" {
			fmt.Fprintf(w, tpl_html,
				charset,
				conf.Name,
				html.EscapeString(outStr))
		} else {
			w.Write([]byte(outStr))
			if(req.req.Referer()!=""){
				w.Write([]byte("<script>window.postMessage && window.parent.postMessage('" + conf.Name + "_height_'+document.body.scrollHeight,'*')</script>"))
			}
		}
	} else if format == "jsonp" {
		w.Header().Set("Content-Type", "text/javascript;charset="+charset)
		cb := r.FormValue("cb")
		if cb == "" {
			cb = fmt.Sprintf("form_%s_jsonp",req.req_path)
		}
		m := make(map[string]string)
		m["data"] = outStr
		jsonByte, _ := json.Marshal(m)
		fmt.Fprintf(w, fmt.Sprintf(`%s(%s)`, cb, string(jsonByte)))
	} else {
		w.Header().Set("Content-Type", "text/plain;charset="+charset)
		w.Write([]byte(outStr))
	}
}
