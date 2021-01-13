package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"code-sync-client/command"
	"code-sync-client/conf"
	"code-sync-client/resource"
)

func main() {
	var confPath string

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&confPath, "conf-path", os.Getenv("HOME")+"/.code-sync-client-conf.json", "client conf path")
	_ = fs.Parse(os.Args[1:])

	confPath = strings.TrimRight(confPath, "/")

	e := conf.Init(confPath)
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(e.Errno())
	}

	e = resource.InitLog("client")
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(e.Errno())
	}
	defer func() {
		resource.FreeLog()
	}()

	fargs := fs.Args()
	if len(fargs) == 0 {
		resource.AccessLogger.Error([]byte("do not has cmd arg"))
		return
	}

	name := strings.TrimSpace(fargs[0])
	cmd := command.NewCommandByName(name)
	if cmd == nil {
		resource.AccessLogger.Error([]byte("unknown cmd: " + name))
		return
	}

	err := cmd.Run(fargs[1:])
	if err != nil {
		resource.AccessLogger.Error([]byte("run command error: " + err.Error()))
	}
}
