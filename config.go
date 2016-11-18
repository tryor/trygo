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
	//指定一个处理Panic异常的函数，如果不指定，将采用默认方式处理
	RecoverFunc func(*Context)

	//响应错误信息方式， HTTP ERROR 或 格式化为json或xml, （默认:false）
	//如果指定了RecoverFunc函数，此配置无效
	ResponseFormatPanic bool

	//RUNMODE_PROD，RUNMODE_DEV
	RunMode int8

	//模板文件位置
	TemplatePath string

	//maximum duration before timing out read of the request, 默认:0(不超时)
	ReadTimeout time.Duration
	//maximum duration before timing out write of the response, 默认:0(不超时)
	WriteTimeout time.Duration

	//如果使用结构体来接收请求参数，可在此设置是否采用域模式传递参数, 默认:false
	//如果值为true, 需要这样传递请求参数：user.account, user为方法参数名(为结构类型)，account为user结构字段名
	FormDomainModel bool
}

const RUNMODE_PROD int8 = 0
const RUNMODE_DEV int8 = 1
