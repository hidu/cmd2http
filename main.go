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
var _port = flag.Int("port", 0, "overwrite the port in the config file")

func main() {
	flag.Parse()

	server := internal.NewServer(*configPath)
	if *_port > 0 {
		server.SetPort(*_port)
	}
	log.Println("exit:", server.Run())
}

// aaa
func init() {
	df := flag.Usage

	flag.Usage = func() {
		df()
		fmt.Fprint(os.Stderr, "\n convert system command as http service\n https://github.com/hidu/cmd2http/\n")
	}
}
