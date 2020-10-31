package conf

import (
	"os"
	"os/user"
)

var BaseConf struct {
	Hostname string
	Username string

	PrjName string
	IsDev   bool

	DataRoot string
}

func initBaseConf() {
	BaseConf.Hostname, _ = os.Hostname()
	curUser, _ := user.Current()
	BaseConf.Username = curUser.Username

	BaseConf.PrjName = ccJson.PrjName
	BaseConf.IsDev = ccJson.IsDev

	BaseConf.DataRoot = ccJson.DataRoot
}
