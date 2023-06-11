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
	"path/filepath"
	"strings"
	"time"
)

type Task struct {
	Ctx     context.Context
	Writer  http.ResponseWriter
	Request *http.Request

	// Path 请求的路径，已经 trim 左右的 /
	//  如 请求 /any ，则 Path = any
	Path string

	srv       *Server
	startTime time.Time
	logInfo   []string
	cacheKey  string
	cmdConf   *cmdConfig
	cmdArgs   []string
	cmdEnv    map[string]string
}

func (tk *Task) Deal() {
	tk.startTime = time.Now()
	tk.logInfo = make([]string, 0, 20)

	tk.log(tk.Request.RemoteAddr, tk.Request.RequestURI)
	defer tk.Close()
	tk.tryExecCmd()
}

func (tk *Task) log(infos ...string) {
	for _, _info := range infos {
		_info = strings.TrimSpace(_info)
		if _info != "" {
			tk.logInfo = append(tk.logInfo, _info)
		}
	}
}

func (tk *Task) Close() {
	used := fmt.Sprintf("time_use:%.3fms", float64(time.Since(tk.startTime).Nanoseconds())/1000000.0)
	log.Println(strings.Join(tk.logInfo, " "), used)
}

var _paramPrefix = "c2h_form_"

func (tk *Task) tryExecCmd() {
	conf, has := tk.srv.config.Commands[tk.Path]
	if !has {
		tk.log("status:404")
		tk.Writer.WriteHeader(404)
		fmt.Fprint(tk.Writer, "<h1>404</h1>")
		return
	}

	args := make([]string, 0, len(conf.paramsAll))
	env := make(map[string]string)

	for _, param := range conf.paramsAll {
		if !param.isValParam {
			args = append(args, param.name)
			continue
		}
		val := tk.Request.FormValue(param.name)
		if val == "" {
			val = tk.Request.PostFormValue(param.name)
		}
		if val == "" {
			val = param.Default
		}

		if param.reg != nil && !param.reg.MatchString(val) {
			tk.log("status:400")
			tk.Writer.WriteHeader(400)
			fmt.Fprintf(tk.Writer, "<h1>400 bad Task param %s=%q</h1>", param.name, val)
		}

		// 特殊的参数：由多个参数合并在一起
		if param.name == "_PARAMS" {
			args = append(args, strings.Fields(val)...)
			continue
		}
		args = append(args, val)
		env[_paramPrefix+param.name] = val
	}

	for k, v := range tk.Request.Form {
		_key := _paramPrefix + k
		_, has := env[_key]
		if !has {
			env[_key] = strings.Join(v, "\t")
		}
	}
	tk.cmdConf = conf
	tk.cmdArgs = args
	tk.cmdEnv = env

	useCache := tk.Request.FormValue("cache")
	// fmt.Println("conf.cache_life",conf.getCacheLife())
	if useCache != "no" && conf.getCacheLife() > 0 {
		tk.cacheKey = GetCacheKey(conf.Command, args)
		//          log.Println("cache_key:",cacheKey)
		cd, err := tk.srv.cache.Get(tk.Request.Context(), tk.cacheKey)
		if err == nil {
			tk.log("cache hit")
			tk.Writer.Header().Add("cache_hit", "1")
			tk.sendResponse(string(cd))
			return
		}
	}
	tk.exec()
}

func (tk *Task) fixArgs(name string, args []string) (string, []string) {
	if name != "exec" || len(args) == 0 {
		return name, args
	}
	return args[0], args[1:]
}

func (tk *Task) exec() {
	conf := tk.cmdConf
	ctx, cancel := context.WithTimeout(tk.Ctx, conf.getTimeout())
	defer cancel()

	name, args := tk.fixArgs(conf.cmdName, tk.cmdArgs)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = filepath.Dir(conf.confDir)
	log.Println("exec:", cmd.String())
	// when use cache,disable the env params
	env := os.Environ()
	if conf.getCacheLife() == 0 {
		for k, v := range tk.cmdEnv {
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
		tk.log("error:" + err.Error())
		tk.Writer.WriteHeader(500)
		fmt.Fprintf(tk.Writer, err.Error()+"\n")
		for argI, argV := range tk.cmdArgs {
			fmt.Fprintf(tk.Writer, "<p>arg[%d]: %q</p>\n", argI, argV)
		}
		return
	}

	err = cmd.Wait()

	isResponseOk := err == nil

	if !isResponseOk || !cmd.ProcessState.Success() {
		tk.Writer.WriteHeader(500)
		tk.Writer.Write([]byte("<h1>Error 500</h1><pre>"))
		tk.Writer.Write([]byte(strings.Join(tk.logInfo, " ")))
		tk.Writer.Write([]byte("\n\nStdOut:\n"))
		tk.Writer.Write(out.Bytes())
		tk.Writer.Write([]byte("\nErrOut:\n"))
		tk.Writer.Write(outErr.Bytes())
		tk.Writer.Write([]byte("</pre>"))
		return
	}

	if out.Len() > 0 && conf.getCacheLife() > 0 {
		tk.srv.cache.Set(tk.Request.Context(), tk.cacheKey, out.Bytes(), conf.getCacheLife())
	}
	tk.sendResponse(out.String())
}

var tplHTML = `<!DOCTYPE html><html><head>
         <meta http-equiv='Content-Type' content='text/html; charset=%s' />
          <title>%s cmd2http</title></head><body><pre>%s</pre></body></html>`

func (tk *Task) sendResponse(outStr string) {
	format := tk.Request.FormValue("format")

	tk.log(fmt.Sprintf("resLen:%d ", len(outStr)))
	w := tk.Writer
	r := tk.Request
	conf := tk.cmdConf
	//      fmt.Println("outStr:",outStr)
	charset := r.FormValue("charset")
	if charset == "" {
		charset = conf.getCharset()
	}
	if format == "" || format == "html" {
		w.Header().Set("Content-Type", "text/html;charset="+charset)
		if format == "" {
			fmt.Fprintf(w, tplHTML, charset, conf.name, html.EscapeString(outStr))
		} else {
			w.Write([]byte(outStr))
			if tk.Request.Referer() != "" {
				w.Write([]byte("<script>window.postMessage && window.parent.postMessage('" + conf.name + "_height_'+document.body.scrollHeight,'*')</script>"))
			}
		}
	} else if format == "jsonp" {
		w.Header().Set("Content-Type", "text/javascript;charset="+charset)
		cb := r.FormValue("cb")
		if cb == "" {
			cb = fmt.Sprintf("form_%s_jsonp", tk.Path)
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
