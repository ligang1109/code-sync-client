package main

import (
	"code-sync-client/command"
	"code-sync-client/conf"
	"code-sync-client/errno"
	"code-sync-client/resource"

	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	var confPath string

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&confPath, "confPath", os.Getenv("HOME")+"/.code-sync-client-conf.json", "client conf path")
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
		fmt.Println("do not has cmd arg")
		os.Exit(errno.ECommonInvalidArg)
	}

	name := strings.TrimSpace(fargs[0])
	cmd := command.NewCommandByName(name)
	if cmd == nil {
		fmt.Println("unknown cmd: " + name)
		os.Exit(errno.ECommonInvalidArg)
	}

	err := cmd.Run(fargs[1:])
	if err != nil {
		fmt.Println("run command error", err)
		os.Exit(errno.ERunCommanError)
	}
}
