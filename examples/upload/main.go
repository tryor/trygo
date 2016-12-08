package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/trygo/ssss"
)

func main() {

	ssss.Register("post", "/upload", &UploadController{}, "Upload")

	ssss.SetStaticPath("/", "static/webcontent/")

	fmt.Println("HTTP ListenAndServe AT ", ssss.DefaultApp.Config.Listen.Addr)
	ssss.Run()

}

type UploadController struct {
	ssss.Controller
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

	lf, err := os.Create(ssss.AppPath + "\\static\\webcontent\\files\\" + fh.Filename)
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
