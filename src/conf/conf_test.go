package conf

import (
	"github.com/goinbox/color"

	"fmt"
	"os"
	"reflect"
	"testing"
)

func init() {
	confPath := os.Getenv("HOME") + "/.code-sync-client-conf.json"

	e := Init(confPath)
	if e != nil {
		fmt.Println("Init error: ", e.Error())
	}
}

func TestConf(t *testing.T) {
	printComplexObjectForTest(&BaseConf)
	printComplexObjectForTest(&LogConf)

	for name, conf := range CodePrjConfMap {
		t.Log("prj", name)
		printComplexObjectForTest(conf)
	}
}

func printComplexObjectForTest(v interface{}) {
	vo := reflect.ValueOf(v)
	elems := vo.Elem()
	ts := elems.Type()

	c := color.Yellow([]byte("Print detail: "))
	fmt.Println(string(c), vo.Type())
	for i := 0; i < elems.NumField(); i++ {
		field := elems.Field(i)
		fmt.Println(ts.Field(i).Name, field.Type(), field.Interface())
	}
}
