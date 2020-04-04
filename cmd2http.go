package main

//go:generate goasset asset.json

import (
	"flag"
	"fmt"
	"os"

	"github.com/hidu/cmd2http/serve"
)

var configPath = flag.String("conf", "./conf/cmd2http.json", "json config file")
var _port = flag.Int("port", 0, "overwrite the port in the config file")

func main() {
	flag.Parse()

	server := serve.NewCmd2HTTPServe(*configPath)
	if *_port > 0 {
		server.SetPort(*_port)
	}
	server.Run()
}

// aaa
func init() {
	df := flag.Usage

	flag.Usage = func() {
		df()
		fmt.Fprintf(os.Stderr, "\n convert system command as http service\n https://github.com/hidu/cmd2http/\n")
	}
}
