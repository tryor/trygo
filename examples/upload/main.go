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

	fmt.Println("HTTP ListenAndServe AT ", ssss.DefaultApp.Config.HttpPort)
	ssss.Run()

}

type UploadController struct {
	ssss.Controller
}

func (c *UploadController) Upload() {
	for _, files := range c.Ctx.Request.MultipartForm.File {
		for _, file := range files {
			saveFile(file)
		}
	}
	c.Ctx.Redirect(301, "/files")
}

func saveFile(fh *multipart.FileHeader) {
	f, err := fh.Open()
	if err != nil {
		ssss.Logger.Error("%v", err)
		return
	}
	defer f.Close()

	lf, err := os.Create(ssss.AppPath + "\\static\\webcontent\\files\\" + fh.Filename)
	if err != nil {
		ssss.Logger.Error("%v", err)
		return
	}
	defer lf.Close()
	_, err = io.Copy(lf, f)
	if err != nil {
		ssss.Logger.Error("%v", err)
		return
	}
}
