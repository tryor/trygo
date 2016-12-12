package main

import (
	"fmt"

	"github.com/tryor/trygo"

	//"github.com/alecthomas/log4go"
	//"github.com/tryor/trygo-log-adaptor/seelog"
)

/**
 * trygo没有功能强大的log服务模块，但可以实现trygo.Logger接口，适配其它log模块,
 * 比如：log4go, seelog等
 */

func main() {
	app := trygo.NewApp()
	//app.Logger = log4go.NewDefaultLogger(log4go.FINEST) //log4go
	//app.Logger, _ = seelog.LoggerFromConfigAsFile("seelog.xml") //seelog

	app.Get("/", func(ctx *trygo.Context) {
		ctx.App.Logger.Info("path:%s", ctx.Request.URL.Path)
		ctx.App.Logger.Info("form:%v", ctx.Request.Form)
		ctx.App.Logger.Warn("test warn")
		ctx.App.Logger.Error("test error")
		ctx.App.Logger.Critical("test critical")
		ctx.App.Logger.Info(123.456, "a", "b", "c")
		ctx.Render("ok")
	})

	fmt.Println("HTTP ListenAndServe AT ", app.Config.Listen.Addr)
	app.Run()
}
