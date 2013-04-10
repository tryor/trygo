package ssss

type Config struct {
	HttpAddr     string
	HttpPort     int
	UseFcgi      bool
	RecoverPanic bool
	RunMode      int8 //0=prod，1=dev
	TemplatePath string
}

const RUNMODE_PROD int8 = 0
const RUNMODE_DEV int8 = 1
