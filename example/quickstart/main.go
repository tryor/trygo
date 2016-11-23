package main

import (
	"fmt"

	"github.com/trygo/ssss"
)

type MainController struct {
	ssss.Controller
}

func (this *MainController) Hello() {
	this.Render("hello world")
}

func (this *MainController) Hi(name string) {
	this.Render("<html><body>Hi, Mr " + name + "</body></html>").Html()
}

func main() {
	ssss.Register("GET|POST", "/", &MainController{}, "Hello")
	ssss.Register("GET|POST", "/hi/(?P<name>[^/]+)$", &MainController{}, "Hi(name string)")

	fmt.Println("HTTP ListenAndServe AT ", ssss.DefaultConfig.HttpPort)
	ssss.Run()

}
