package conf

type logConfJson struct {
	Level          int `json:"level"`
	AsyncQueueSize int `json:"async_queue_size"`
	Bufsize        int `json:"bufsize"`
}

var LogConf struct {
	RootPath       string
	Level          int
	AsyncQueueSize int
	Bufsize        int
}

func initLogConf() {
	LogConf.RootPath = BaseConf.DataRoot + "/logs"
	LogConf.Level = ccJson.Log.Level
	LogConf.AsyncQueueSize = ccJson.Log.AsyncQueueSize
	LogConf.Bufsize = ccJson.Log.Bufsize
}
