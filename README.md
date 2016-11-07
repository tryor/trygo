## SSSS
=======
ssss 是基于Golang的http、web服务框架。部份思路和源码来自于github.com/astaxie/beego。此框架的目标并不是想做一个大而全的web容器，它主要用于开发底层高性能高可靠性的http服务。支持如下特性：MVC,类型内方法路由、正则路由,JSON/JSON(JQueryCallback)/XML服务，模板，静态文件输出。暂时不支持会话管理模块。

ssss HTTP and WEB services of framework for Golang。It is mainly used to develop the underlying HTTP service,Support feature:MVC,Methods the routing and regular routing,JSON/JSON(JQueryCallback)/XML service,template,Static file output。Temporarily does not support session management module。

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
	cfg.RecoverPanic = true
	cfg.TemplatePath = "static/templates"

	app := ssss.NewApp(&cfg)

	app.Register("GET|POST", "/e", &MainController{}, "Example", "p1, p2 string")
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
	HttpAddr     string
	HttpPort     int
	UseFcgi      bool
	PrintPanic   bool
	RunMode      int8 //0=prod，1=dev
	TemplatePath string
	ReadTimeout  time.Duration // maximum duration before timing out read of the request, 默认:0(不超时)
	WriteTimeout time.Duration // maximum duration before timing out write of the response, 默认:0(不超时)	
}

const RUNMODE_PROD int8 = 0
const RUNMODE_DEV int8 = 1
```

## Thank End
============