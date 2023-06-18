// Copyright(C) 2023 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2023/6/5

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var server = flag.String("s", os.Getenv("Ch_Remote_Server"), "server addr, default value from env 'Ch_Remote_Server'")
var timeout = flag.Int("t", 60, "timeout, unit is seconds")
var debug = flag.Bool("d", os.Getenv("Ch_Remote_Debug") == "yes", "print debug logs")

func main() {
	flag.Parse()
	if *server == "" {
		log.Fatalln("server address is empty")
	}
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("args required")
	}

	u, err := url.Parse(*server)
	if err != nil {
		log.Fatalln(err.Error())
	}
	user := u.User
	u.User = nil

	api := u.String()

	params := url.Values{}
	params.Add("_PARAMS", strings.Join(args, " "))
	params.Add("format", "plain")
	bf := bytes.NewBufferString(params.Encode())
	req, err := http.NewRequest(http.MethodPost, api, bf)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if user != nil {
		req.Header.Set("AK", user.Username())
		tm := strconv.FormatInt(time.Now().Unix(), 10)
		req.Header.Set("TM", tm)
		psw, _ := user.Password()
		req.Header.Set("TK", tk(user.Username(), tm, psw))
	}
	ch := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}
	resp, err := ch.Do(req)
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

func tk(user string, tm string, psw string) string {
	m5 := md5.New()
	m5.Write([]byte(user))
	m5.Write([]byte(tm))
	m5.Write([]byte(psw))
	return hex.EncodeToString(m5.Sum(nil))
}
