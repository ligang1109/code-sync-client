package svc

import (
	"code-sync-client/conf"
	"code-sync-client/resource"

	"os"
)

func init() {
	_ = conf.Init(os.Getenv("GOPATH"))

	_ = resource.InitLog("test")

}
