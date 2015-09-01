package main

import (
	"flag"
	"fmt"
	"github.com/hidu/cmd2http/serve"
	"os"
)

var configPath = flag.String("conf", "./conf/cmd2http.json", "json config file")
var _port = flag.Int("port", 0, "http server port,overwrite the port in the config file")
var _help = flag.Bool("help", false, "show help")

func main() {
	flag.Parse()
	if *_help {
		printHelp()
		os.Exit(0)
	}

	server := serve.NewCmd2HTTPServe(*configPath)
	if *_port > 0 {
		server.SetPort(*_port)
	}
	server.Run()
}

func printHelp() {
	fmt.Println("useage:")
	flag.PrintDefaults()
	fmt.Println("\nconfig demo:\n")
	fmt.Println(string(serve.LoadRes("res/conf/cmd2http.conf")))
}
