package main

//go:generate goasset asset.json

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hidu/cmd2http/internal"
)

var configPath = flag.String("conf", "./conf/cmd2http.json", "json config file")
var port = flag.Int("port", 0, "overwrite the port in the config file")
var auth = flag.String("auth", "", "overwrite the BasicAuth in the config file")

func main() {
	flag.Parse()

	server := internal.NewServer(*configPath)
	if *port > 0 {
		server.SetPort(*port)
	}
	if *auth != "" {
		server.SetBasicAuth(*auth)
	}

	log.Println("exit:", server.Run())
}

func init() {
	df := flag.Usage

	flag.Usage = func() {
		df()
		fmt.Fprint(os.Stderr, "\n convert CLI as HTTP service\n https://github.com/hidu/cmd2http/\n")
	}
}
