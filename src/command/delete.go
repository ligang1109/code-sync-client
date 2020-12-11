package command

import (
	"code-sync-client/conf"
	syncSvc "code-sync-client/svc/sync"
	"errors"
	"github.com/goinbox/gomisc"
)

const (
	CmdNameDelete = "delete"
)

func init() {
	register(CmdNameDelete, newDeleteCommand)
}

func newDeleteCommand() ICommand {
	dc := &deleteCommand{
		baseCommand: NewBaseCommand(),
	}

	dc.AddMustHaveArgs("prj-name").
		SetRunFunc(dc.run)

	dc.Fs.StringVar(&dc.prjName, "prj-name", "", "prj-name")

	return dc
}

type deleteCommand struct {
	*baseCommand

	prjName string
	rpath   string

	cpc *conf.CodePrjConf
}

func (dc *deleteCommand) run() error {
	if err := dc.parseArgs(); err != nil {
		return err
	}

	ss := syncSvc.NewSyncSvc([]byte(dc.prjName + " sync file"))

	return ss.DeleteFile(dc.cpc, dc.rpath)
}

func (dc *deleteCommand) parseArgs() error {
	var ok bool
	dc.cpc, ok = conf.CodePrjConfMap[dc.prjName]
	if !ok {
		return errors.New("prj " + dc.prjName + "not exist")
	}

	if len(dc.Fs.Args()) == 0 {
		return errors.New("do not have arg relative path")
	}

	dc.rpath = dc.Fs.Arg(0)
	apath := dc.cpc.PrjHome + "/" + dc.rpath
	if !gomisc.FileExist(apath) {
		return errors.New("apath " + apath + " not exist")
	}

	return nil
}
