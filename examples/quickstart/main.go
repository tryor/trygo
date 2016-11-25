package main

import (
	"fmt"

	"github.com/trygo/ssss"
)

func main() {

	ssss.Get("/", func(ctx *ssss.Context) {
		ctx.Render("hello world")
	})

	fmt.Println("HTTP ListenAndServe AT ", ssss.DefaultApp.Config.HttpPort)
	ssss.Run()

}
