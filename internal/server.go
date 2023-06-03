package internal

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/fsgo/fsgo/fsfs"
	"github.com/hidu/goutils/cache"
)

type Server struct {
	logPath string
	config  *Config
	Cache   cache.Cache
}

func NewServer(confPath string) *Server {
	server := new(Server)
	server.config = loadConfig(confPath)
	return server
}

// SetPort set cmd server http port
func (srv *Server) SetPort(port int) {
	srv.config.Port = port
}

// Run start http server
func (srv *Server) Run() error {
	static, err := fs.Sub(resourceWeb, "resource/static")
	if err != nil {
		return err
	}
	http.Handle("/s/", http.FileServer(http.Dir("./")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	http.Handle("/favicon.ico", http.FileServer(http.FS(favicon)))
	http.HandleFunc("/help", srv.handlerHelp)
	http.HandleFunc("/", srv.index)

	addr := fmt.Sprintf(":%d", srv.config.getPort())
	log.Println("listen at:", addr)

	srv.setupCache()
	srv.setupLog()

	return http.ListenAndServe(addr, nil)
}

func (srv *Server) setupLog() {
	if srv.logPath == "" {
		return
	}
	logFile := &fsfs.Rotator{
		ExtRule: "1hour",
		Path:    srv.logPath,
	}
	log.SetOutput(logFile)
}

func (srv *Server) setupCache() {
	if srv.cacheAble() {
		srv.Cache = cache.NewFileCache(srv.config.CacheDir)
		log.Println("use file cache,cache dir:", srv.config.CacheDir)
	} else {
		srv.Cache = cache.NewNoneCache()
		log.Print("use none cache")
	}
	srv.Cache.StartGcTimer(600)
}

func (srv *Server) cacheAble() bool {
	return len(srv.config.CacheDir) > 5
}
