package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	//"github.com/tryor/trygo"
	//"github.com/tryor/trygo-bridge/fasthttp"
	//"github.com/tryor/trygo-bridge/graceful"

	"tryor/trygo"
)

/**
 * 演示所有监听模式或自定义监听模式
 */

func main() {

	go ListenAndServe(7086) //Default http
	//go ListenAndServe(9081, &trygo.HttpServer{Network: "tcp4"}) //Default http, tcp4
	//go ListenAndServe(6086, &trygo.FcgiHttpServer{})            //Fcgi
	//go ListenAndServe(7087, &trygo.FcgiHttpServer{Network: "unix"})                                //Fcgi unix
	//go ListenAndServe(4433, &trygo.TLSHttpServer{CertFile: "cert.pem", KeyFile: "key.pem"})        //Https
	//go ListenAndServe(9090, &fasthttp.FasthttpServer{})                                            //FastHttp
	//go ListenAndServe(4439, &fasthttp.TLSFasthttpServer{CertFile: "cert.pem", KeyFile: "key.pem"}) //FastHttps
	//go ListenAndServe(9060, &graceful.GracefulServer{Timeout: 10 * time.Second})                   //graceful

	select {}

}

func ListenAndServe(port int, server ...trygo.Server) {
	app := trygo.NewApp()
	app.Config.Listen.Addr = fmt.Sprintf("0.0.0.0:%v", port)
	//	app.Config.Listen.ReadTimeout = time.Second * 2
	//	app.Config.Listen.WriteTimeout = time.Second * 3
	//app.Config.Listen.Concurrency = 10
	//app.Config.MaxRequestBodySize = 1024 * 1024 * 8
	//app.Config.AutoParseRequest = false

	app.Get("/statinfo", func(ctx *trygo.Context) {
		type Statinfo struct {
			ConcurrentConns     int32
			PeakConcurrentConns int32
		}
		var statinfo Statinfo
		statinfo.ConcurrentConns = app.Statinfo.ConcurrentConns()
		statinfo.PeakConcurrentConns = app.Statinfo.PeakConcurrentConns()
		ctx.Render(statinfo).Json()
	})

	app.Post("/reqinfo", func(ctx *trygo.Context) {
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
		ck, err := req.Cookie("Cookie2")
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
			Cookie(&http.Cookie{Name: "Cookie1", Value: "1", Domain: "127.0.0.1", MaxAge: 100, Expires: time.Now().Add(100 * time.Second), HttpOnly: true}).
			Cookie(&http.Cookie{Name: "Cookie2", Value: "2"})
	})

	app.SetStaticPath("/", "static/webcontent/")

	fmt.Println("ListenAndServe AT ", port)
	app.Run(server...)
}
