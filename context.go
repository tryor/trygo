package ssss

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Multipart      bool
	Form           url.Values
}

func (ctx *Context) Abort(status int, body string) {
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte(body))
}

func (ctx *Context) Redirect(status int, url_ string) {
	ctx.ResponseWriter.Header().Set("Location", url_)
	ctx.ResponseWriter.WriteHeader(status)
}

func (ctx *Context) NotModified() {
	ctx.ResponseWriter.WriteHeader(304)
}

func (ctx *Context) NotFound(message string) {
	ctx.Error(404, message)
}

func (ctx *Context) Error(code int, message string) (err error) {
	ctx.ResponseWriter.WriteHeader(code)
	_, err = ctx.ResponseWriter.Write([]byte(message))
	return
}

func (ctx *Context) ContentType(typ string) {
	ext := typ
	if !strings.HasPrefix(typ, ".") {
		ext = "." + typ
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		ctx.ResponseWriter.Header().Set("Content-Type", ctype)
	} else {
		ctx.ResponseWriter.Header().Set("Content-Type", typ)
	}
}

func (ctx *Context) AddHeader(hdr string, val string) {
	ctx.ResponseWriter.Header().Add(hdr, val)
}

//func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
func (ctx *Context) SetHeader(hdr string, val string) {
	ctx.ResponseWriter.Header().Set(hdr, val)
}

//func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
//	if unique {
//		ctx.ResponseWriter.Header().Set(hdr, val)
//	} else {
//		ctx.ResponseWriter.Header().Add(hdr, val)
//	}
//}

//Sets a cookie -- duration is the amount of time in seconds. 0 = forever
func (ctx *Context) SetCookie(name string, value string, age int64) {
	var utctime time.Time
	if age == 0 {
		// 2^31 - 1 seconds (roughly 2038)
		utctime = time.Unix(2147483647, 0)
	} else {
		utctime = time.Unix(time.Now().Unix()+age, 0)
	}
	cookie := fmt.Sprintf("%s=%s; expires=%s", name, value, webTime(utctime))
	ctx.AddHeader("Set-Cookie", cookie)
}

func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}
