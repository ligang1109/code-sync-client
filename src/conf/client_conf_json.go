package conf

import (
	"github.com/goinbox/gomisc"
)

var ccJson clientConfJson

type clientConfJson struct {
	PrjName string `json:"prj_name"`
	IsDev   bool   `json:"is_dev"`

	DataRoot string `json:"data_root"`

	Log logConfJson `json:"log"`

	CodePrjMap map[string]*codePrjConfJson `json:"code_prj_map"`
}

func initClientConfJson(path string) error {
	err := gomisc.ParseJsonFile(path, &ccJson)
	if err != nil {
		return err
	}

	return nil
}
