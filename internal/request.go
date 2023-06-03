package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type request struct {
	writer    http.ResponseWriter
	req       *http.Request
	reqPath   string
	cmd2      *Cmd2HttpServe
	startTime time.Time
	logInfo   []string
	stop      bool
	cacheKey  string
	cmdConf   *cmdItem
	cmdArgs   []string
	cmdEnv    map[string]string
}

func (req *request) Deal() {
	req.startTime = time.Now()
	req.logInfo = make([]string, 0, 20)
	req.reqPath = strings.Trim(req.req.URL.Path, "/")

	//    fmt.Println("path:",req.reqPath)

	req.log(req.req.RemoteAddr, req.req.RequestURI)
	defer req.Close()
	req.handleStatic()
	if req.stop {
		return
	}
	req.tryExecCmd()
}

func (req *request) log(infos ...string) {
	for _, _info := range infos {
		_info = strings.TrimSpace(_info)
		if _info != "" {
			req.logInfo = append(req.logInfo, _info)
		}
	}
}

func (req *request) Close() {
	used := fmt.Sprintf("time_use:%.3fms", float64(time.Since(req.startTime).Nanoseconds())/1000000.0)
	log.Println(strings.Join(req.logInfo, " "), used)
}

func (req *request) handleStatic() {
	if req.reqPath == "" {
		if IsFileExists("./s/index.html") {
			http.Redirect(req.writer, req.req, "/s/", http.StatusFound)
		} else {
			req.cmd2.myHandlerHelp(req.writer, req.req)
		}
		req.stop = true
	}
}

var _paramPrefix = "c2h_form_"

func (req *request) tryExecCmd() {
	conf, has := req.cmd2.config.Cmds[req.reqPath]
	if !has {
		req.log("status:404")
		req.writer.WriteHeader(404)
		fmt.Fprint(req.writer, "<h1>404</h1>")
		return
	}

	args := make([]string, 0, len(conf.paramsAll))
	env := make(map[string]string)

	for _, param := range conf.paramsAll {
		if !param.isValParam {
			args = append(args, param.Name)
			continue
		}
		val := req.req.FormValue(param.Name)
		if val == "" {
			val = param.DefaultValue
		}

		// 特殊的参数：由多个参数合并在一起
		if param.Name == "_PARAMS" {
			args = append(args, strings.Fields(val)...)
			continue
		}
		args = append(args, val)
		env[_paramPrefix+param.Name] = val
	}

	for k, v := range req.req.Form {
		_key := _paramPrefix + k
		_, has := env[_key]
		if !has {
			env[_key] = strings.Join(v, "\t")
		}
	}
	req.cmdConf = conf
	req.cmdArgs = args
	req.cmdEnv = env

	useCache := req.req.FormValue("cache")
	//      fmt.Println("conf.cache_life",conf.cache_life)
	if useCache != "no" && conf.CacheLife > 3 {
		req.cacheKey = GetCacheKey(conf.CmdRaw, args)
		//          log.Println("cache_key:",cacheKey)
		cacheHas, cacheData := req.cmd2.Cache.Get(req.cacheKey)
		if cacheHas {
			req.log("cache hit")
			req.writer.Header().Add("cache_hit", "1")
			req.sendResponse(string(cacheData))
			return
		}
	}
	req.exec()
}

func (req *request) exec() {
	conf := req.cmdConf

	ctx, cancel := context.WithTimeout(req.req.Context(), conf.getTimeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, conf.Cmd, req.cmdArgs...)
	log.Println("exec:", cmd.String())
	// when use cache,disable the env params
	env := os.Environ()
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
		for argI, argV := range req.cmdArgs {
			fmt.Fprintf(req.writer, "arg_%d:%s\n", argI, argV)
		}
		return
	}

	err = cmd.Wait()

	isResponseOk := err == nil

	if !isResponseOk || !cmd.ProcessState.Success() {
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

var tplHTML = `<!DOCTYPE html><html><head>
         <meta http-equiv='Content-Type' content='text/html; charset=%s' />
          <title>%s cmd2http</title></head><body><pre>%s</pre></body></html>`

func (req *request) sendResponse(outStr string) {
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
			fmt.Fprintf(w, tplHTML, charset, conf.Name, html.EscapeString(outStr))
		} else {
			w.Write([]byte(outStr))
			if req.req.Referer() != "" {
				w.Write([]byte("<script>window.postMessage && window.parent.postMessage('" + conf.Name + "_height_'+document.body.scrollHeight,'*')</script>"))
			}
		}
	} else if format == "jsonp" {
		w.Header().Set("Content-Type", "text/javascript;charset="+charset)
		cb := r.FormValue("cb")
		if cb == "" {
			cb = fmt.Sprintf("form_%s_jsonp", req.reqPath)
		}
		m := make(map[string]string)
		m["data"] = outStr
		jsonByte, _ := json.Marshal(m)
		fmt.Fprintf(w, `%s(%s)`, cb, string(jsonByte))
	} else {
		w.Header().Set("Content-Type", "text/plain;charset="+charset)
		w.Write([]byte(outStr))
	}
}
