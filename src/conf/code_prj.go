package conf

import "strings"

type codeSyncServerConfJson struct {
	Host  string `json:"host"`
	Port  string `json:"port"`
	Token string `json:"token"`
}

type codePrjConfJson struct {
	PrjHome            string                    `json:"prj_home"`
	ExcludeList        []string                  `json:"exclude_list"`
	CodeSyncServerList []*codeSyncServerConfJson `json:"code_sync_server_list"`
}

type CodeSyncServerConf struct {
	Host  string
	Port  string
	Token string
}

type CodePrjConf struct {
	PrjName     string
	PrjHome     string
	ExcludeList []string

	CodeSyncServerList []*CodeSyncServerConf
}

var CodePrjConfMap map[string]*CodePrjConf

func initCodePrjConf() {
	CodePrjConfMap = make(map[string]*CodePrjConf)
	for name, item := range ccJson.CodePrjMap {
		cpc := &CodePrjConf{
			PrjName:     name,
			PrjHome:     strings.TrimRight(item.PrjHome, "/"),
			ExcludeList: item.ExcludeList,
		}

		for _, server := range item.CodeSyncServerList {
			cpc.CodeSyncServerList = append(cpc.CodeSyncServerList, &CodeSyncServerConf{
				Host:  server.Host,
				Port:  server.Port,
				Token: server.Token,
			})
		}

		CodePrjConfMap[name] = cpc
	}
}
