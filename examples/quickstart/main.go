package main

import (
	"fmt"

	"github.com/tryor/trygo"
)

func main() {

	trygo.Get("/", func(ctx *trygo.Context) {
		ctx.Render("hello world")
	})

	fmt.Println("HTTP ListenAndServe AT ", trygo.DefaultApp.Config.Listen.Addr)
	trygo.Run()

}
