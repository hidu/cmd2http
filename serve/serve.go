package serve

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hidu/goutils/cache"
	"github.com/hidu/goutils/fs"
	"github.com/hidu/goutils/time_util"
)

// Cmd2HttpServe server struct
type Cmd2HttpServe struct {
	logFile   *os.File
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
func (cmd2 *Cmd2HttpServe) Run() {
	cmd2.setupCache()

	http.Handle("/s/", http.FileServer(http.Dir("./")))
	http.Handle("/res/", Asset.HTTPHandler("/"))
	http.Handle("/favicon.ico", Asset.FileHandlerFunc("/res/css/favicon.ico"))
	http.HandleFunc("/help", cmd2.myHandlerHelp)
	http.HandleFunc("/", cmd2.myHandlerRoot)

	addr := fmt.Sprintf(":%d", cmd2.config.Port)
	log.Println("listen at", addr)
	cmd2.setupLog()
	defer cmd2.logFile.Close()

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println(err.Error())
		log.Println(err.Error())
	}
}
func (cmd2 *Cmd2HttpServe) setupLog() {
	cmd2.logFile, _ = os.OpenFile(cmd2.logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	log.SetOutput(cmd2.logFile)

	time_util.SetInterval(func() {
		if !fs.FileExists(cmd2.logPath) {
			cmd2.logFile.Close()
			cmd2.logFile, _ = os.OpenFile(cmd2.logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
			log.SetOutput(cmd2.logFile)
		}
	}, 30)
}

func (cmd2 *Cmd2HttpServe) setupCache() {
	if len(cmd2.config.CacheDir) > 5 {
		cmd2.Cache = cache.NewFileCache(cmd2.config.CacheDir)
		log.Println("use file cache,cache dir:", cmd2.config.CacheDir)
		cmd2.cacheAble = true
	} else {
		cmd2.Cache = cache.NewNoneCache()
		log.Printf("use none cache")
	}
	cmd2.Cache.StartGcTimer(600)
}
