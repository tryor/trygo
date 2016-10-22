package ssss

import (
	"time"
)

type Config struct {
	HttpAddr string
	HttpPort int
	UseFcgi  bool
	//是否打印Panic详细信息
	PrintPanic bool
	RunMode    int8 //0=prod，1=dev
	//模板文件位置
	TemplatePath string
	ReadTimeout  time.Duration // maximum duration before timing out read of the request, 默认:0(不超时)
	WriteTimeout time.Duration // maximum duration before timing out write of the response, 默认:0(不超时)
}

const RUNMODE_PROD int8 = 0
const RUNMODE_DEV int8 = 1
