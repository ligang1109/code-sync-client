package command

import (
	"errors"
	"github.com/goinbox/gomisc"

	"code-sync-client/conf"
	syncSvc "code-sync-client/svc/sync"
)

const (
	CmdNameUpload = "upload"
)

func init() {
	register(CmdNameUpload, newUploadCommand)
}

func newUploadCommand() ICommand {
	uc := &uploadCommand{
		baseCommand: NewBaseCommand(),
	}

	uc.AddMustHaveArgs("prj-name").
		SetRunFunc(uc.run)

	uc.Fs.StringVar(&uc.prjName, "prj-name", "", "prj-name")

	return uc
}

type uploadCommand struct {
	*baseCommand

	prjName string
	rpath   string

	cpc *conf.CodePrjConf
}

func (uc *uploadCommand) run() error {
	if err := uc.parseArgs(); err != nil {
		return err
	}

	ss := syncSvc.NewSyncSvc([]byte(uc.prjName + " sync file"))
	return ss.UploadFile(uc.cpc, uc.rpath)
}

func (uc *uploadCommand) parseArgs() error {
	var ok bool
	uc.cpc, ok = conf.CodePrjConfMap[uc.prjName]
	if !ok {
		return errors.New("prj " + uc.prjName + "not exist")
	}

	if len(uc.Fs.Args()) == 0 {
		return errors.New("do not have arg relative path")
	}

	uc.rpath = uc.Fs.Arg(0)
	apath := uc.cpc.PrjHome + "/" + uc.rpath
	if !gomisc.FileExist(apath) {
		return errors.New("apath " + apath + " not exist")
	}

	return nil
}
