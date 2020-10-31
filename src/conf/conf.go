package conf

import (
	"github.com/goinbox/exception"
	"github.com/goinbox/gomisc"

	"code-sync-client/errno"
)

func Init(path string) *exception.Exception {
	if !gomisc.FileExist(path) {
		return exception.New(errno.ESysInvalidConfPath, "confPath not exists")
	}

	err := initClientConfJson(path)
	if err != nil {
		return exception.New(errno.ESysInitClientConfFail, "init clientConfJson error: "+err.Error())
	}

	initBaseConf()
	initLogConf()
	initCodePrjConf()

	return nil
}
