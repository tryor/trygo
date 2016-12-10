package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tryor/fasthttp-trygo-bridging"
	"github.com/tryor/trygo"

	//	"tryor/fasthttp-trygo-bridging"
	//	"tryor/trygo"
)

/**
 * 演示所有监听模式或自定义监听模式
 */

func main() {

	go ListenAndServe(9080)                                                                //Default http
	go ListenAndServe(9081, &ssss.HttpServer{Network: "tcp4"})                             //Default http, tcp4
	go ListenAndServe(7086, &ssss.FcgiHttpServer{})                                        //Fcgi
	go ListenAndServe(4433, &ssss.TLSHttpServer{CertFile: "cert.pem", KeyFile: "key.pem"}) //Https
	//	go ListenAndServe(9090, &bridging.FasthttpServer{})                 //FastHttp
	//go ListenAndServe(4439, &bridging.TLSFasthttpServer{CertFile: "cert.pem", KeyFile: "key.pem"}) //FastHttps
	select {}

}

func ListenAndServe(port int, server ...ssss.Server) {
	app := ssss.NewApp()
	app.Config.Listen.Addr = fmt.Sprintf("0.0.0.0:%v", port)
	//app.Config.Listen.Concurrency = 10
	//app.Config.MaxRequestBodySize = 1024 * 1024 * 8
	//app.Config.AutoParseRequest = false

	app.Post("/reqinfo", func(ctx *ssss.Context) {
		reqinfo := ""
		req := ctx.Request

		form := req.Form
		reqinfo += fmt.Sprintf("ctx.Request.Form() => %v\n", form)

		postForm := req.PostForm
		reqinfo += fmt.Sprintf("ctx.Request.PostForm() => %v\n", postForm)

		mform := req.MultipartForm
		reqinfo += fmt.Sprintf("ctx.Request.MultipartForm() => %v\n", mform)

		username, password, ok := req.BasicAuth()
		reqinfo += fmt.Sprintf("req.BasicAuth() => username:%v, password:%v, ok:%v\n", username, password, ok)
		reqinfo += fmt.Sprintf("req.Closed() => %v\n", req.Close)
		reqinfo += fmt.Sprintf("req.ContentLength() => %v\n", req.ContentLength)
		reqinfo += fmt.Sprintf("req.Header().Get() Content-Type => %v\n", req.Header.Get("Content-Type"))
		ck, err := req.Cookie("ckname11")
		reqinfo += fmt.Sprintf("req.Cookie() => %v, err:%v\n", ck, err)
		reqinfo += fmt.Sprintf("req.Cookies() => %v\n", req.Cookies())

		file, _, err := req.FormFile("file1")
		if file != nil {
			defer file.Close()
		}
		reqinfo += fmt.Sprintf("req.FormFile() => %v, err:%v\n", file, err)
		reqinfo += fmt.Sprintf("req.FormValue().p1 => %v\n", req.FormValue("p1"))
		reqinfo += fmt.Sprintf("req.PostFormValue().p2 => %v\n", req.PostFormValue("p2"))
		reqinfo += fmt.Sprintf("req.Host => %v\n", req.Host)
		reqinfo += fmt.Sprintf("req.Method => %v\n", req.Method)
		reqinfo += fmt.Sprintf("req.Proto => %v\n", req.Proto)
		reqinfo += fmt.Sprintf("req.ProtoMajor => %v\n", req.ProtoMajor)
		reqinfo += fmt.Sprintf("req.ProtoMinor => %v\n", req.ProtoMinor)
		reqinfo += fmt.Sprintf("req.ProtoAtLeast => %v\n", req.ProtoAtLeast(req.ProtoMajor, req.ProtoMinor))
		reqinfo += fmt.Sprintf("req.Referer => %v\n", req.Referer())
		reqinfo += fmt.Sprintf("req.RemoteAddr => %v\n", req.RemoteAddr)
		reqinfo += fmt.Sprintf("req.RequestURI => %v\n", req.RequestURI)
		reqinfo += fmt.Sprintf("req.TLS => %v\n", req.TLS)
		reqinfo += fmt.Sprintf("req.TransferEncoding => %v\n", req.TransferEncoding)
		reqinfo += fmt.Sprintf("req.URL => %v\n", req.URL)
		reqinfo += fmt.Sprintf("req.UserAgent => %v\n", req.UserAgent())

		body := req.Body
		reqinfo += fmt.Sprintf("req.Body() => %v\n", body)
		if body != nil {
			bodybuf, err := ioutil.ReadAll(body)
			if err != nil {
				app.Logger.Error("%v", err)
			} else {
				reqinfo += fmt.Sprintf("req.Body().Data(%v) => %v\n", len(bodybuf), string(bodybuf))
			}
		}
		ctx.Render(reqinfo).
			Cookie(&http.Cookie{Name: "ckame211", Value: "ckval211", Domain: "127.0.0.1", MaxAge: 100, Expires: time.Now().Add(100 * time.Second), HttpOnly: true}).
			Cookie(&http.Cookie{Name: "ckname22", Value: "ckval22"})
	})

	app.SetStaticPath("/", "static/webcontent/")

	fmt.Println("ListenAndServe AT ", port)
	app.Run(server...)
}
