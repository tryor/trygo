## SSSS
=======
ssss 是基于Golang的http、web服务框架。部份思路和源码来自于github.com/astaxie/beego。此框架的目标并不是想做一个大而全的web容器，它主要用于开发底层http服务。没有会话管理模块，不支持正则路由。支持如下特性：MVC,方法路由，JSON/JSON(JQueryCallback)/XML服务，模板，静态文件输出。

ssss HTTP and WEB services of framework for Golang。It is mainly used to develop the underlying HTTP service,No session management module, does not support the regular route。Support feature:MVC,Methods the routing,JSON/JSON(JQueryCallback)/XML service,template,Static file output。

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
ssss.Register("GET", "/", &MainController{}, "Func1")
ssss.Register("POST", "/", &MainController{}, "Func2")
ssss.Register("GET|POST", "/", &MainController{}, "Func3")
ssss.Register("PUT", "/", &MainController{}, "Func4")
ssss.Register("*", "/", &MainController{}, "Func5")
```


## Static files
============
ssss.SetStaticPath("/", "static/webroot/")



