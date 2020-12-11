package command

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/goinbox/gomisc"

	"code-sync-client/conf"
	"code-sync-client/resource"
	"code-sync-client/misc"
	syncSvc "code-sync-client/svc/sync"
)

const (
	CmdNameAuto = "auto"

	SyncOpUpload = "upload"
	SyncOpDelete = "delete"

	NotifyFileBufsize  = 256
	SyncMaxWaitSeconds = 3
)

func init() {
	register(CmdNameAuto, newAutoCommand)
}

func newAutoCommand() ICommand {
	ac := &autoCommand{
		baseCommand: NewBaseCommand(),
		nfCh:        make(chan *nofityFile, NotifyFileBufsize),
	}

	ac.SetRunFunc(ac.run)

	return ac
}

type nofityFile struct {
	op    string
	cpc   *conf.CodePrjConf
	rpath string
}

func (nf *nofityFile) String() string {
	return "op:" + nf.op + " prj: " + nf.cpc.PrjName + " rpath: " + nf.rpath
}

type autoCommand struct {
	*baseCommand

	nfCh chan *nofityFile
}

func (ac *autoCommand) run() error {
	go ac.readNotifyFileRoutine()

	wg := new(sync.WaitGroup)

	for prj, cpc := range conf.CodePrjConfMap {
		resource.AccessLogger.Info([]byte("watch prj " + prj))
		wg.Add(1)
		go ac.watchPrjRoutine(wg, cpc)
	}

	wg.Wait()

	resource.AccessLogger.Info([]byte("all watchRrjRoutine return"))

	return nil
}

func (ac *autoCommand) readNotifyFileRoutine() {
	notifyFileMap := make(map[string]*nofityFile)
	ticker := time.NewTicker(time.Second * SyncMaxWaitSeconds)
	ss := syncSvc.NewSyncSvc([]byte("auto sync"))

	for {
		select {
		case nf := <-ac.nfCh:
			resource.AccessLogger.Info([]byte("read notify file " + nf.String()))
			key := nf.cpc.PrjName + "|" + nf.rpath
			notifyFileMap[key] = nf
		case <-ticker.C:
			for _, nf := range notifyFileMap {
				switch nf.op {
				case SyncOpUpload:
					_ = ss.UploadFile(nf.cpc, nf.rpath)
				case SyncOpDelete:
					_ = ss.DeleteFile(nf.cpc, nf.rpath)
				default:
					resource.AccessLogger.Error([]byte("unknown op"))
				}
			}
			notifyFileMap = make(map[string]*nofityFile)
		}
	}
}

func (ac *autoCommand) watchPrjRoutine(wg *sync.WaitGroup, cpc *conf.CodePrjConf) {
	defer wg.Done()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		resource.AccessLogger.Error([]byte("watch prj " + cpc.PrjName + " new wather error: " + err.Error()))
		return
	}
	defer func() {
		_ = watcher.Close()
	}()

	err = ac.watchDirRecursive(cpc.PrjHome, watcher, cpc)
	if err != nil {
		return
	}

	err = ac.readPrjEvents(watcher, cpc)
	if err != nil {
		return
	}
}

func (ac *autoCommand) watchDirRecursive(path string, watcher *fsnotify.Watcher, cpc *conf.CodePrjConf) error {
	dirs, err := ac.listDirsInDir(path)
	if err != nil {
		resource.AccessLogger.Error([]byte("watch prj " + cpc.PrjName + " list dirs error: " + err.Error()))
		return err
	}

	for _, dir := range dirs {
		rpath := misc.RelativePath(cpc.PrjHome, dir)
		if misc.PathInExcludeList(rpath, cpc.ExcludeList) {
			resource.AccessLogger.Info([]byte("watch prj " + cpc.PrjName + " watcher exclude dir: " + dir))
			continue
		}

		resource.AccessLogger.Info([]byte("watch prj " + cpc.PrjName + " watcher add dir: " + dir))
		err = watcher.Add(dir)
		if err != nil {
			resource.AccessLogger.Error([]byte("watch prj " + cpc.PrjName + " watcher add error: " + err.Error()))
			return err
		}
	}

	return nil
}

func (ac *autoCommand) listDirsInDir(rootDir string) ([]string, error) {
	rootDir = strings.TrimRight(rootDir, "/")
	if !gomisc.DirExist(rootDir) {
		return nil, errors.New("Dir " + rootDir + " not exists")
	}

	dirList := []string{rootDir}

	for i := 0; i < len(dirList); i++ {
		curDir := dirList[i]
		file, err := os.Open(dirList[i])
		if err != nil {
			return nil, err
		}

		fis, err := file.Readdir(-1)
		if err != nil {
			return nil, err
		}

		for _, fi := range fis {
			path := curDir + "/" + fi.Name()
			if fi.IsDir() {
				dirList = append(dirList, path)
			}
		}
	}

	return dirList, nil
}

func (ac *autoCommand) readPrjEvents(watcher *fsnotify.Watcher, cpc *conf.CodePrjConf) error {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				resource.AccessLogger.Error([]byte("watch prj " + cpc.PrjName + " read event channel close"))
				return nil
			}
			resource.AccessLogger.Debug([]byte("watch prj " + cpc.PrjName + " read event: " + event.String()))
			ac.processPrjEvents(&event, watcher, cpc)
		case err, ok := <-watcher.Errors:
			if !ok {
				resource.AccessLogger.Error([]byte("watch prj " + cpc.PrjName + " read event channel close"))
				return nil
			}
			resource.AccessLogger.Error([]byte("watch prj " + cpc.PrjName + " read event error: " + err.Error()))
		}
	}
}

func (ac *autoCommand) processPrjEvents(event *fsnotify.Event, watcher *fsnotify.Watcher, cpc *conf.CodePrjConf) {
	apath := event.Name
	rpath := misc.RelativePath(cpc.PrjHome, apath)

	if misc.PathInExcludeList(rpath, cpc.ExcludeList) {
		resource.AccessLogger.Info([]byte("watch prj " + cpc.PrjName + " notify exclude rpath: " + rpath))
		return
	}

	//delete dir REMOVE|WRITE
	//move dir REMOVE|RENAME

	if event.Op&fsnotify.Remove == fsnotify.Remove {
		ac.notify(SyncOpDelete, rpath, cpc)
		return
	}

	if event.Op&fsnotify.Rename == fsnotify.Rename {
		ac.notify(SyncOpDelete, rpath, cpc)
		return
	}

	if event.Op&fsnotify.Create == fsnotify.Create {
		if gomisc.DirExist(apath) {
			_ = ac.watchDirRecursive(apath, watcher, cpc)
		}
		ac.notify(SyncOpUpload, rpath, cpc)
		return
	}

	if event.Op&fsnotify.Write == fsnotify.Write {
		ac.notify(SyncOpUpload, rpath, cpc)
		return
	}
}

func (ac *autoCommand) notify(op, rpath string, cpc *conf.CodePrjConf) {
	ac.nfCh <- &nofityFile{
		op:    op,
		cpc:   cpc,
		rpath: rpath,
	}
}
