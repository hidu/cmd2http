package internal

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/hidu/goutils/cache"
)

// Cmd2HttpServe server struct
type Cmd2HttpServe struct {
	logPath   string
	config    *serverConf
	Cache     cache.Cache
	cacheAble bool
}

// NewCmd2HTTPServe load cmd server
func NewCmd2HTTPServe(confPath string) *Cmd2HttpServe {
	server := new(Cmd2HttpServe)
	server.config = loadConfig(confPath)
	return server
}

// SetPort set cmd server http port
func (cmd2 *Cmd2HttpServe) SetPort(port int) {
	cmd2.config.Port = port
}

// Run start http server
func (cmd2 *Cmd2HttpServe) Run() error {
	cmd2.setupCache()

	static, err := fs.Sub(resourceWeb, "resource/static")
	if err != nil {
		return err
	}
	http.Handle("/s/", http.FileServer(http.Dir("./")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	http.Handle("/favicon.ico", http.FileServer(http.FS(favicon)))
	http.HandleFunc("/help", cmd2.myHandlerHelp)
	http.HandleFunc("/", cmd2.index)

	addr := fmt.Sprintf(":%d", cmd2.config.Port)
	log.Println("listen at", addr)
	cmd2.setupLog()

	return http.ListenAndServe(addr, nil)
}

func (cmd2 *Cmd2HttpServe) setupLog() {
	if cmd2.logPath == "" {
		return
	}
	logFile := &fsfs.Rotator{
		ExtRule: "1hour",
		Path:    cmd2.logPath,
	}
	log.SetOutput(logFile)
}

func (cmd2 *Cmd2HttpServe) setupCache() {
	if len(cmd2.config.CacheDir) > 5 {
		cmd2.Cache = cache.NewFileCache(cmd2.config.CacheDir)
		log.Println("use file cache,cache dir:", cmd2.config.CacheDir)
		cmd2.cacheAble = true
	} else {
		cmd2.Cache = cache.NewNoneCache()
		log.Print("use none cache")
	}
	cmd2.Cache.StartGcTimer(600)
}
