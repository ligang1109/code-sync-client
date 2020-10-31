package errno

const (
	Success = 0

	ESysInvalidConfPath    = 11
	ESysInitClientConfFail = 12
	ESysInitLogFail        = 13
	ESysSavePidFileFail    = 14
	ESysMysqlError         = 15
	ESysRedisError         = 16
	ESysMongoError         = 17

	ECommonFileNotExist    = 101
	ECommonReadFileError   = 102
	ECommonJsonEncodeError = 103
	ECommonJsonDecodeError = 104
	ECommonInvalidArg      = 105

	ERunCommanError = 1001
)
