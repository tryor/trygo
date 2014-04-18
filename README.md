## SSSS
=======
ssss 是基于Golang的http、web服务框架。部份思路和源码来自于github.com/astaxie/beego。此框架的目标并不是想做一个大而全的web容器，它主要用于开发底层高性能高可靠性的http服务。支持如下特性：MVC,类型内方法路由，JSON/JSON(JQueryCallback)/XML服务，模板，静态文件输出。暂时不支持会话管理模块和正则路由。

ssss HTTP and WEB services of framework for Golang。It is mainly used to develop the underlying HTTP service,Support feature:MVC,Methods the routing,JSON/JSON(JQueryCallback)/XML service,template,Static file output。Temporarily does not support session management module and a regular routing。

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
	//cfg.HttpAddr = "0.0.0.0"
	cfg.HttpPort = 8080
	ssss.Run(&cfg)
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
	RecoverPanic bool
	RunMode      int8 //0=prod，1=dev
	TemplatePath string
}

const RUNMODE_PROD int8 = 0
const RUNMODE_DEV int8 = 1
```

## Thank End
============