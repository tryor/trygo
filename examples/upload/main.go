package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/tryor/trygo"
)

func main() {
	trygo.DefaultApp.Config.MaxRequestBodySize = 1024 * 1024 * 20

	trygo.Register("post", "/upload", &UploadController{}, "Upload")

	trygo.SetStaticPath("/", "static/webcontent/")

	fmt.Println("HTTP ListenAndServe AT ", trygo.DefaultApp.Config.Listen.Addr)
	trygo.Run()

}

type UploadController struct {
	trygo.Controller
}

func (c *UploadController) Upload() {
	mform := c.Ctx.Request.MultipartForm
	for _, files := range mform.File {
		for _, file := range files {
			c.saveFile(file)
		}
	}
	c.Ctx.Redirect(302, "/files")
}

func (c *UploadController) saveFile(fh *multipart.FileHeader) {
	f, err := fh.Open()
	if err != nil {
		c.App.Logger.Error("%v", err)
		return
	}
	defer f.Close()

	lf, err := os.Create(trygo.AppPath + "\\static\\webcontent\\files\\" + fh.Filename)
	if err != nil {
		c.App.Logger.Error("%v", err)
		return
	}
	defer lf.Close()
	_, err = io.Copy(lf, f)
	if err != nil {
		c.App.Logger.Error("%v", err)
		return
	}
}
