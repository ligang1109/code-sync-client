package conf

import (
	"os"
	"os/user"
)

var CommonConf struct {
	Hostname string
	Username string

	PrjName string
	IsDev   bool

	DataRoot string
}

func initBaseConf() {
	CommonConf.Hostname, _ = os.Hostname()
	curUser, _ := user.Current()
	CommonConf.Username = curUser.Username

	CommonConf.PrjName = ccJson.PrjName
	CommonConf.IsDev = ccJson.IsDev

	CommonConf.DataRoot = ccJson.DataRoot
}
