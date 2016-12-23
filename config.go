package trygo

import (
	"time"
)

type config struct {
	Listen listenConfig

	//生产或开发模式，值：PROD, DEV
	RunMode int8

	//模板文件位置
	TemplatePath string

	//请求主体数据量大小限制, 默认：defaultMaxRequestBodySize
	MaxRequestBodySize int64

	//是否自动分析请求参数，默认:true
	AutoParseRequest bool

	//如果使用结构体来接收请求参数，可在此设置是否采用域模式传递参数, 默认:false
	//如果值为true, 需要这样传递请求参数：user.account, user为方法参数名(为结构类型)，account为user结构字段名
	FormDomainModel bool

	//指示绑定请求参数时发生错误是否抛出异常, 默认:true
	//如果设置为false, 当绑定数据出错时，将采用相应类型的默认值填充数据，并返回error
	ThrowBindParamPanic bool

	//指定一个处理Panic异常的函数，如果不指定，将采用默认方式处理
	RecoverFunc func(*Context)
	//是否打印Panic详细信息, 开发模式肯定会打印, @see defaultRecoverFunc
	//如果是自定义RecoverFunc，PrintPanicDetail配置将无效
	//默认:true
	PrintPanicDetail bool

	//打开统计信息功能
	StatinfoEnable bool

	Render renderConfig
}

type listenConfig struct {
	//listen addr, format: "[ip]:<port>", ":7086", "0.0.0.0:7086", "127.0.0.1:7086"
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	//并发连接的最大数目, 默认：defaultConcurrency
	Concurrency int
	//连接Keep-Alive时间限制， 默认0, 无限制
	//MaxKeepaliveDuration time.Duration
}

type renderConfig struct {

	//是否自动从请求参数中解析响应数据格式
	//如果被设置，将从请求参数中自动获取的FormatParamName参数以及JsoncallbackParamName参数值
	//默认:false
	AutoParseFormat bool

	//默认：fmt
	FormatParamName string
	//默认: jsoncb
	JsoncallbackParamName string

	//默认是否使用Result结构对结果进行包装， @see result.go
	//如果设置此参数，将默认设置Render.Wrap()
	//当Render.Wrap()后，如果不设置响应数据格式，将默认为:json
	//默认:false
	Wrap bool

	//默认是否对响应数据进行gzip压缩
	//默认:false
	Gzip bool
}

func newConfig() *config {
	cfg := &config{}

	cfg.RunMode = PROD
	cfg.TemplatePath = ""

	cfg.MaxRequestBodySize = defaultMaxRequestBodySize
	cfg.AutoParseRequest = true
	cfg.FormDomainModel = false
	cfg.ThrowBindParamPanic = true

	cfg.RecoverFunc = defaultRecoverFunc
	cfg.PrintPanicDetail = true
	cfg.StatinfoEnable = false

	cfg.Listen.Addr = "0.0.0.0:7086"
	cfg.Listen.ReadTimeout = 0
	cfg.Listen.WriteTimeout = 0
	cfg.Listen.Concurrency = defaultConcurrency
	//cfg.Listen.MaxKeepaliveDuration = 0

	cfg.Render.AutoParseFormat = false
	cfg.Render.FormatParamName = "fmt"
	cfg.Render.JsoncallbackParamName = "jsoncb"
	cfg.Render.Wrap = false
	cfg.Render.Gzip = false
	return cfg
}

//生产或开发模式
const (
	PROD = iota
	DEV
)

//数据渲染格式
const (
	FORMAT_JSON = "json"
	FORMAT_XML  = "xml"
	FORMAT_TXT  = "txt"
	FORMAT_HTML = "html"
	// other ...
)

const defaultMaxRequestBodySize = 4 * 1024 * 1024

const defaultConcurrency = 256 * 1024
