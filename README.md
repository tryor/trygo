## SSSS
=======
ssss 是基于Golang的http、web服务框架。此框架的目标并不是想做一个大而全的web服务容器，它主要用于开发底层高性能高可靠性的http服务。支持如下特性：MVC,类型内方法路由、正则路由,JSON/JSON(JQueryCallback)/XML结果响应支持，模板，静态文件输出。暂时不支持会话管理模块。

ssss HTTP and WEB services of framework for Golang. It is mainly used to develop the underlying HTTP service, Support feature:MVC,Methods the routing and regular routing,JSON/JSON(JQueryCallback)/XML result response support,template,Static file output。Temporarily does not support session management module。

ssss is licensed under the Apache Licence, Version 2.0
(http://www.apache.org/licenses/LICENSE-2.0.html).

## Installation
============
To install:

    go get -u github.com/trygo/ssss

## Quick Start
============
Here is the canonical "Hello, world" example app for ssss:

```go
package main

import (
    "github.com/trygo/ssss"
)

type MainController struct {
    ssss.Controller
}

func (this *MainController) Hello() {
    this.RenderText("hello world")
}

func (this *MainController) Hi() {
    this.RenderHtml("<html><body>Hi</body></html>")
}

func main() {
    ssss.Register("GET|POST", "/", &MainController{}, "Hello")
    ssss.Register("GET|POST", "/hi", &MainController{}, "Hi")

    var cfg ssss.Config
    cfg.HttpPort = 8080
    ssss.Run(&cfg)
}
```
A better understanding of the SSSS example:
```go
package main

import (
	"fmt"
	"github.com/trygo/ssss"
	"runtime"
	//"trygo/ssss"
)

type MainController struct {
	ssss.Controller
}

type ResultData struct {
	Hello string  `json:"hello" xml:"hello"`
	Val1  int     `json:"val1" xml:"val1"`
	Val2  bool    `json:"val2" xml:"val2"`
	Val3  float64 `json:"val3" xml:"val3"`
	Val4  string  `json:"val4,omitempty" xml:"val4,omitempty"`
	Val5  string  `json:"val5" xml:"val5"`
}

func (this *MainController) Example(p1, p2 string) {

	fmt.Printf("p1=%v\n", p1)
	fmt.Printf("p2=%v\n", p2)

	var rs ResultData
	rs.Hello = "hello world"
	rs.Val1 = 100
	rs.Val2 = true
	rs.Val3 = float64(100.001)
	rs.Val4 = p1
	rs.Val5 = p2

	this.RenderSucceed("json", &rs)
}

func (this *MainController) Example1() {
	form, err := this.ParseForm()
	if err != nil {
		this.RenderError(this.Ctx.Request.FormValue("fmt"), err)
	}

	p1 := form.Get("p1")
	p2 := form.Get("p2")
	fmt.Print("p1=", p1)
	fmt.Print("p2=", p2)

	var rs ResultData
	rs.Hello = "hello world"
	rs.Val1 = 100
	rs.Val2 = true
	rs.Val3 = float64(100.001)
	this.RenderSucceed(form.Get("fmt"), &rs)
}

func (this *MainController) Example2() {
	form, err := this.ParseForm()
	if err != nil {
		errrs := ssss.NewErrorResult(ssss.ERROR_CODE_RUNTIME, fmt.Sprintf("parse parameter error, %v", err))
		this.RenderError(this.Ctx.Request.FormValue("fmt"), errrs)
	}
	var rsdata ResultData
	rsdata.Hello = "hello world"
	rsdata.Val1 = 100
	rsdata.Val2 = true
	rsdata.Val3 = float64(100.001)

	sucrs := ssss.NewSucceedResult(&rsdata)
	this.RenderSucceed(form.Get("fmt"), sucrs)
}

func (this *MainController) Example3() {
	this.TplNames = "admin/index.tpl" //The actual position(cfg.TemplatePath+this.TplNames)：static/templates/admin/index.tp
	this.Data["nodes"] = nil          //service.GetNodes()
	this.Data["insts"] = nil          //service.GetInstances()
	this.RenderTemplate()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	go ListenAndServe(9080)
	//go ListenAndServe(9081)

	fmt.Println("Test: http://127.0.0.1:9080/e")
	fmt.Println("Test: http://127.0.0.1:9080/e1")
	fmt.Println("Test: http://127.0.0.1:9080/e1?fmt=json")
	fmt.Println("Test: http://127.0.0.1:9080/e2")
	fmt.Println("Test: http://127.0.0.1:9080/e2?fmt=xml")
	fmt.Println("Test: http://127.0.0.1:9080/e3")
	select {}
}

func ListenAndServe(port int) {
	var cfg ssss.Config
	cfg.HttpPort = port
	cfg.PrintPanic = true
	cfg.TemplatePath = "static/templates"

	app := ssss.NewApp(&cfg)

	app.Register("GET|POST", "/e", &MainController{}, "Example(p1, p2 string)")
	app.Register("GET|POST", "/e1", &MainController{}, "Example1")
	app.Register("GET|POST", "/e2", &MainController{}, "Example2")
	app.Register("GET|POST", "/e3", &MainController{}, "Example3")

	app.SetStaticPath("/", "static/webcontent/")

	fmt.Println("HTTP ListenAndServe AT ", port)
	app.Run()
}

```

## Router
============
```go
ssss.Register("GET", "/f1", &MainController{}, "Func1")
ssss.Register("POST", "/f2", &MainController{}, "Func2")
ssss.Register("GET|POST", "/f3", &MainController{}, "Func3")
ssss.Register("PUT", "/f4", &MainController{}, "Func4")
ssss.Register("GET|POST", "/admin/login", &AdminController{}, "Login")
ssss.Register("*", "/admin/index", &AdminController{}, "Index")

ssss.RegisterPattern("GET|POST", "/.*", &MainController{}, "Func5")
ssss.RegisterPattern("GET|POST", "/admin/.*", &AdminController{}, "Index")

```

## Request
============
```go

Http handler method parameter is struct, the struct field tag name is `field`,
tag attributes will have name,limit,scope,default,require,pattern, for example:
`field:"name,limit:20,scope:[1 2 3],default:1,require,pattern:xxxxx"`
scope: [1 2 3] or [1~100] or [0~] or [~0] or [100~] or [~100] or [~-100 -20~-10 -1 0 1 2 3 10~20 100~]

type UserForm struct {
	Account string `field:"account,limit:20,require"` 
	Pwd     string `field:"pwd,limit:10,require"`
	Name    string `field:"name,limit:20"`
	Sex     int    `field:"sex,scope:[1 2 3],default:1"` 
	Age     uint   `field:"age,scope:[0~200]"` 
	Email   string `field:"email,limit:30,pattern:\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*"` 
	Photo   string
}


type MainController struct {
	ssss.Controller
}
func (this *MainController) Create(userform UserForm) {
	ssss.Logger.Info("user=%v", user)
	//...
	user := service.UserService.Create(userform)
	//...
	this.RenderSucceed("json", user)
}

ssss.Register("GET|POST", "/user/create", &MainController{}, "Create(userform UserForm)")



```
```go
Http handler method parameter is base data type, support parameter tag.

const (
	accountTag = `account:"limit:20,require"`
	pwdTag     = `pwd:"limit:20,require"`
)

var LoginTags = []string{accountTag, pwdTag}

func (this *MainController) Login(account, pwd string) {

	fmt.Printf("account=%v\n", account)
	fmt.Printf("pwd=%v\n", pwd)

	this.RenderSucceed("json", "sessionid")
}


ssss.Register("GET|POST", "/user/login", &MainController{}, "Login(account, pwd string)", LoginTags...)


```
## Static files
============
ssss.SetStaticPath("/", "static/webroot/")

## Render
============

All the default render:

```go
func (c *Controller) Render(contentType string, data []byte)
func (c *Controller) RenderHtml(content string)
func (c *Controller) RenderText(content string)
func (c *Controller) RenderJson(data interface{})
func (c *Controller) RenderJQueryCallback(jsoncallback string, data interface{})
func (c *Controller) RenderXml(data interface{})
func (c *Controller) RenderTemplate(contentType ...string)

func (c *Controller) RenderSucceed(fmt string, data interface{})
func (c *Controller) RenderError(fmt string, err interface{})

```

## View / Template
============

template view path set

```go
var cfg ssss.Config
cfg.TemplatePath = "static/templates"
```
template names

ssss will find the template from cfg.TemplatePath. the file is set by user like：
```go
c.TplNames = "admin/add.tpl"
```
then ssss will find the file in the path:static/templates/admin/add.tpl

if you don't set TplNames,sss will find like this:
```go
c.TplNames = c.ChildName + "/" + c.MethodName + "." + c.TplExt
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
type Config struct {
	HttpAddr string
	HttpPort int
	UseFcgi  bool
	
	//是否打印Panic详细信息
	PrintPanic bool
	
	//响应错误信息方式， HTTP ERROR 或 格式化为json或xml, （默认:false）
	ResponseFormatPanic bool
	
	//RUNMODE_PROD，RUNMODE_DEV
	RunMode             int8 
	
	//模板文件位置
	TemplatePath string
	
	//maximum duration before timing out read of the request, 默认:0(不超时)
	ReadTimeout  time.Duration 
	//maximum duration before timing out write of the response, 默认:0(不超时)
	WriteTimeout time.Duration 

	//如果使用结构体来接收请求参数，可在此设置是否采用域模式传递参数, 默认:false
	//如果值为true, 需要这样传递请求参数：user.account, user为方法参数名(为结构类型)，account为user结构字段名
	FormDomainModel bool
}

const RUNMODE_PROD int8 = 0
const RUNMODE_DEV int8 = 1
```

## Thank End
============