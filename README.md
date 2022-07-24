## trygo
=======
trygo 是基于Golang的http、web服务框架。此框架的目标并不是想做一个大而全的web服务容器，它主要用于开发底层高性能高可靠性的http服务。支持如下特性：RESTful,MVC,类型内方法路由、正则路由,JSON/JSON(JQueryCallback)/XML结果响应支持，模板，静态文件输出，net.Listener过滤，http.Handler过滤。暂时不支持会话管理模块。

trygo HTTP and WEB services of framework for Golang. It is mainly used to develop the underlying HTTP service, Support feature:RESTful,MVC,Methods the routing and regular routing,JSON/JSON(JQueryCallback)/XML result response support,template,Static file output, net.Listener filter, http.Handler filter. Temporarily does not support session management module.

trygo is licensed under the Apache Licence, Version 2.0
(http://www.apache.org/licenses/LICENSE-2.0.html).

## Installation
============
To install:

    go get -u github.com/tryor/trygo

## Quick Start
============
Here is the canonical "Hello, world" example app for trygo:

```go
package main

import (
	"fmt"

	"github.com/tryor/trygo"
)

func main() {

	trygo.Get("/", func(ctx *trygo.Context) {
		ctx.Render("hello world")
	})

	fmt.Println("HTTP ListenAndServe AT ", trygo.DefaultApp.Config.HttpPort)
	trygo.Run()

}
```
A better understanding of the trygo example:


@see (https://github.com/tryor/trygo/tree/master/examples)


## Router
============
```go

trygo.Register(method string, path string, c IController, name string, params ...string)
trygo.RegisterHandler(pattern string, h http.Handler)
trygo.RegisterRESTful(pattern string, c IController)
trygo.RegisterFunc(methods, pattern string, f HandlerFunc)
trygo.Get(pattern string, f HandlerFunc)
trygo.Post(pattern string, f HandlerFunc) 
trygo.Put(pattern string, f HandlerFunc)
 ...
 ```

for example： 
@see (https://github.com/tryor/trygo/tree/master/examples/router)



## Request
============
```go

Http handler method parameter is struct, the struct field tag name is `param`,
tag attributes will have name,limit,scope,default,require,pattern,layout for example:
`param:"name,limit:20,scope:[1 2 3],default:1,require,pattern:xxxxx"`
scope: [1 2 3] or [1~100] or [0~] or [~0] or [100~] or [~100] or [~-100 -20~-10 -1 0 1 2 3 10~20 100~]

type UserForm struct {
	Account  string `param:"account,limit:20,require"` 
	Pwd      string `param:"pwd,limit:10,require"`
	Name     string `param:"name,limit:20"`
	Sex      int    `param:"sex,scope:[1 2 3],default:1"` 
	Age      uint   `param:"age,scope:[0~200]"` 
	Birthday time.Time `param:"birthday,layout:2006-01-02|2006-01-02 15:04:05"` 
	Email    string `param:"email,limit:30,pattern:\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*"` 
	Photo    string
}


type MainController struct {
	trygo.Controller
}
func (this *MainController) Create(userform UserForm) {
	trygo.Logger.Info("user=%v", user)
	//...
	user := service.UserService.Create(userform)
	//...
	this.Render(user)
}

trygo.Register("POST", "/user/create", &MainController{}, "Create(userform UserForm)")



```
```go
Http handler method parameter is base data type, support parameter tag.

const (
	accountTag = `param:"account,limit:20,require"`
	pwdTag     = `param:"pwd,limit:20,require"`
)

var LoginTags = []string{accountTag, pwdTag}

func (this *MainController) Login(account, pwd string) {

	fmt.Printf("account=%v\n", account)
	fmt.Printf("pwd=%v\n", pwd)

	this.Render("sessionid")
}


trygo.Register("POST", "/user/login", &MainController{}, "Login(account, pwd string)", LoginTags...)


```

## Render
============
All the default render:


@see (https://github.com/tryor/trygo/tree/master/examples/render)


## Static files
============
trygo.SetStaticPath("/", "static/webroot/")


## View / Template
============

template view path set

```go
trygo.SetViewsPath("static/templates/")
```
template names

trygo will find the template from cfg.TemplatePath. the file is set by user like：
```go
c.TplNames = "admin/add.tpl"
```
then trygo will find the file in the path:static/templates/admin/add.tpl

if you don't set TplNames,sss will find like this:
```go
c.TplNames = c.ChildName + "/" + c.ActionName + "." + c.TplExt
```

render template

```go
c.TplNames = "admin/add.tpl"
c.Data["data"] = you data
c.RenderTemplate()
```

## Config
============
```go

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

	Render renderConfig
}

type listenConfig struct {
	//listen addr, format: "[ip]:<port>", ":7086", "0.0.0.0:7086", "127.0.0.1:7086"
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	//并发连接的最大数目, 默认：defaultConcurrency
	Concurrency int
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

```

## Thank End
=============
