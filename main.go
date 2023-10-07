package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hidu/cmd2http/internal"
)

var configPath = flag.String("conf", "./conf/app.toml", "app config file")
var listen = flag.String("listen", "", "overwrite the Listen Addr in the config file")
var users = flag.String("users", "", `overwrite the Users in the config file.
e.g.: 
    user1:psw1             --> Only one user
    user1:psw1;user2:psw2  --> Two users
    no                     --> No login required
`)

func main() {
	flag.Parse()
	server := internal.NewServer(*configPath)
	if *listen != "" {
		server.SetListen(*listen)
	}
	if *users != "" {
		server.SetUsers(*users)
	}

	log.Println("exit:", server.Run())
}

func init() {
	df := flag.Usage

	flag.Usage = func() {
		df()
		fmt.Fprint(os.Stderr, `
convert CLI as HTTP service
Site: https://github.com/hidu/cmd2http/
`)
	}
}
