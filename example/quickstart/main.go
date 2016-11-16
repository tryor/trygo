package main

import (
	"fmt"

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

	fmt.Println("HTTP ListenAndServe AT ", cfg.HttpPort)
	ssss.Run(&cfg)

}
