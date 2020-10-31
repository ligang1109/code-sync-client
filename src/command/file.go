package command

import (
	"code-sync-client/conf"
	"code-sync-client/svc/upload"
	"errors"
	"github.com/goinbox/gomisc"
)

const (
	CmdNameFile = "file"
)

func init() {
	register(CmdNameFile, newFileCommand)
}

func newFileCommand() ICommand {
	fc := &FileCommand{
		baseCommand: NewBaseCommand(),
	}

	fc.AddMustHaveArgs("prj-name").
		SetRunFunc(fc.run)

	fc.Fs.StringVar(&fc.prjName, "prj-name", "", "prj-name")

	return fc
}

type FileCommand struct {
	*baseCommand

	prjName string
	rpath   string

	cpc *conf.CodePrjConf
}

func (fc *FileCommand) run() error {
	if err := fc.parseArgs(); err != nil {
		return err
	}

	us := upload.NewUploadSvc([]byte(fc.prjName + "|" + fc.rpath))

	return us.UploadFile(fc.cpc, fc.rpath, FileVersionIgnore)
}

func (fc *FileCommand) parseArgs() error {
	var ok bool
	fc.cpc, ok = conf.CodePrjConfMap[fc.prjName]
	if !ok {
		return errors.New("prj " + fc.prjName + "not exist")
	}

	if len(fc.Fs.Args()) == 0 {
		return errors.New("do not have arg relative path")
	}

	fc.rpath = fc.Fs.Arg(0)
	apath := fc.cpc.PrjHome + "/" + fc.rpath
	if !gomisc.FileExist(apath) {
		return errors.New("apath " + apath + " not exist")
	}

	return nil
}
