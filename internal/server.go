package internal

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/fsgo/fscache"
	"github.com/fsgo/fscache/filecache"
	"github.com/fsgo/fscache/nopcache"
	"github.com/fsgo/fsgo/fsfs"
)

type Server struct {
	config *Config
	cache  fscache.Cache
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

func (srv *Server) SetBasicAuth(auth string) {
	srv.config.BasicAuth = auth
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
	lp := srv.config.getLogPath()
	if lp == "" {
		return
	}
	logFile := &fsfs.Rotator{
		ExtRule: "1hour",
		Path:    lp,
	}
	log.SetOutput(logFile)
}

func (srv *Server) setupCache() {
	if srv.cacheAble() {
		opt := &filecache.Option{
			Dir: srv.config.getCacheDir(),
		}
		var err error
		srv.cache, err = filecache.New(opt)
		if err != nil {
			log.Fatalln("init file cache failed:", err)
		}
		log.Println("use file cache,cache dir:", srv.config.CacheDir)
	} else {
		srv.cache = nopcache.Nop
		log.Print("use none cache")
	}
}

func (srv *Server) cacheAble() bool {
	return len(srv.config.CacheDir) > 5
}
