// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/5

package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var server = flag.String("s", os.Getenv("Ch_Remote_Server"), "server addr, default value from env 'Ch_Remote_Server'")
var timeout = flag.Int("t", 60, "timeout, unit is seconds")
var debug = flag.Bool("d", true, "print debug logs")

func main() {
	flag.Parse()
	if *server == "" {
		log.Fatalln("server address is empty")
	}
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("args required")
	}

	params := url.Values{}
	params.Add("_PARAMS", strings.Join(args, " "))
	params.Add("format", "plain")
	api := *server + "?" + params.Encode()
	ch := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}
	resp, err := ch.Get(api)
	if *debug {
		log.Println("call:", api)
	}
	if err != nil {
		log.Fatalln(err.Error())
	}

	if *debug {
		log.Println("Response Status:", resp.StatusCode)
	}
	defer resp.Body.Close()
	n, err1 := io.Copy(os.Stdout, resp.Body)
	if err1 != nil {
		log.Fatalln("read Response Body failed, n=", n, "err=", err)
	}
}
