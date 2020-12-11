package conf

import (
	"github.com/goinbox/goerror"
	"github.com/goinbox/gomisc"

	"code-sync-client/errno"
)

func Init(path string) *goerror.Error {
	if !gomisc.FileExist(path) {
		return goerror.New(errno.ESysInvalidConfPath, "confPath not exists")
	}

	err := initClientConfJson(path)
	if err != nil {
		return goerror.New(errno.ESysInitClientConfFail, "init clientConfJson error: "+err.Error())
	}

	initBaseConf()
	initLogConf()
	initCodePrjConf()

	return nil
}
