package conf

import "strings"

type codeSyncServerConfJson struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type codePrjConfJson struct {
	PrjHome        string                  `json:"prj_home"`
	Token          string                  `json:"token"`
	CodeSyncServer *codeSyncServerConfJson `json:"code_sync_server"`
}

type CodeSyncServerConf struct {
	Host string
	Port string
}

type CodePrjConf struct {
	PrjName string
	PrjHome string
	Token   string

	CodeSyncServer *CodeSyncServerConf
}

var CodePrjConfMap map[string]*CodePrjConf

func initCodePrjConf() {
	CodePrjConfMap = make(map[string]*CodePrjConf)
	for name, item := range ccJson.CodePrjMap {
		CodePrjConfMap[name] = &CodePrjConf{
			PrjName: name,
			PrjHome: strings.TrimRight(item.PrjHome, "/"),
			Token:   item.Token,
			CodeSyncServer: &CodeSyncServerConf{
				Host: item.CodeSyncServer.Host,
				Port: item.CodeSyncServer.Port,
			},
		}
	}
}
