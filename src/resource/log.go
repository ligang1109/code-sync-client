package resource

import (
	"code-sync-client/conf"
	"code-sync-client/errno"
	"github.com/goinbox/goerror"

	"github.com/goinbox/golog"
)

var accessLogWriter golog.IWriter
var AccessLogger golog.ILogger

func InitLog(systemName string) *goerror.Error {
	if conf.CommonConf.IsDev {
		accessLogWriter = golog.NewConsoleWriter()
	} else {
		fw, err := golog.NewFileWriter(conf.LogConf.RootPath+"/"+systemName+"_access.log", conf.LogConf.Bufsize)
		if err != nil {
			return goerror.New(errno.ESysInitLogFail, err.Error())
		}
		accessLogWriter = golog.NewAsyncWriter(fw, conf.LogConf.AsyncQueueSize)
	}
	AccessLogger = NewLogger(accessLogWriter)

	return nil
}

func NewLogger(writer golog.IWriter) golog.ILogger {
	return golog.NewSimpleLogger(writer, golog.NewSimpleFormater()).SetLogLevel(conf.LogConf.Level)
}

func FreeLog() {
	accessLogWriter.Free()
}
