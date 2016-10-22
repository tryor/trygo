package ssss

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type IController interface {
	Init(app *App, ct *Context, cn string, mn string)
	//如果返回false, 将终止此请求
	Prepare() bool
	Finish()
}

type Controller struct {
	Data       map[interface{}]interface{}
	Ctx        *Context
	ChildName  string
	MethodName string
	TplNames   string
	TplExt     string
	App        *App
	//panic抛出的异常，默认为nil, 如果不想继续抛出异常， 设置c.PanicInfo = nil
	PanicInfo interface{}
}

func (c *Controller) Init(app *App, ctx *Context, cn string, mn string) {
	c.Data = make(map[interface{}]interface{})
	c.App = app
	c.ChildName = cn
	c.MethodName = mn
	c.Ctx = ctx
	//c.TplNames = ""
	c.TplExt = "tpl"
}

func (c *Controller) Prepare() bool {
	return true
}

func (c *Controller) Finish() {
}

func (c *Controller) Redirect(url string, code int) {
	c.Ctx.Redirect(code, url)
}

func (c *Controller) Error(code int, message string) {
	c.Ctx.Error(code, message)
}

func (c *Controller) Render(contentType string, data []byte) (err error) {
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(data)))
	c.Ctx.ContentType(contentType)
	_, err = c.Ctx.ResponseWriter.Write(data)
	return
}

func (c *Controller) RenderHtml(content string) (err error) {
	return c.Render("html", []byte(content))
}

func (c *Controller) RenderText(content string) (err error) {
	return c.Render("txt", []byte(content))
}

func (c *Controller) RenderJson(data interface{}) (err error) {
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	return c.Render("application/json", content)
}

func (c *Controller) RenderJQueryCallback(jsoncallback string, data interface{}) (e error) {
	var content []byte
	switch data.(type) {
	case string:
		content = []byte(data.(string))
	case []byte:
		content = data.([]byte)
	default:
		var err error
		content, err = json.Marshal(data)
		if err != nil {
			http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	bjson := []byte(jsoncallback)
	bjson = append(bjson, '(')
	bjson = append(bjson, content...)
	bjson = append(bjson, ')')
	return c.Render("application/json", bjson)
}

func (c *Controller) RenderXml(data interface{}) (err error) {
	content, err := xml.Marshal(data)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	return c.Render("xml", content)
}

func (c *Controller) RenderTemplate(contentType ...string) (err error) {
	if c.TplNames == "" {
		c.TplNames = c.ChildName + "/" + c.MethodName + "." + c.TplExt
	}
	_, file := path.Split(c.TplNames)
	subdir := path.Dir(c.TplNames)
	ibytes := bytes.NewBufferString("")

	if c.App.config.RunMode == RUNMODE_DEV {
		c.App.buildTemplate()
	}

	t := c.App.TemplateRegistor.Templates[subdir]
	if t == nil {
		http.Error(c.Ctx.ResponseWriter, "Internal Server Error (template not exist)", http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(ibytes, file, c.Data)
	if err != nil {
		log.Println("template Execute err:", err)
	}
	icontent, _ := ioutil.ReadAll(ibytes)
	if len(contentType) > 0 {
		return c.Render(contentType[0], icontent)
	} else {
		return c.Render("html", icontent)
	}
}

//fmt值指示响应结果格式，当前支持:json或xml, 默认为:json
//如果是json格式结果，支持jsoncallback
func (c *Controller) RenderError(fmt string, err interface{}) (e error) {
	//fmt := c.Ctx.Request.FormValue("fmt")
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
		return c.renderJsonError(err)
	case "xml":
		return c.renderXmlError(err)
	default:
		return c.renderJsonError(err)
	}
}

//fmt值指示响应结果格式，当前支持:json或xml, 默认为:json
//如果是json格式结果，支持jsoncallback
func (c *Controller) RenderSucceed(fmt string, data interface{}) (err error) {
	//fmt := c.Ctx.Request.FormValue("fmt")
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
		return c.renderJsonSucceed(data)
	case "xml":
		return c.renderXmlSucceed(data)
	default:
		return c.renderJsonSucceed(data)
	}
}

func (c *Controller) renderJsonError(err interface{}) (e error) {
	//log.Fatal(err)
	rs := ConvertErrorResult(err)

	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	if jsoncallback != "" {
		return c.RenderJQueryCallback(jsoncallback, rs)
	} else {
		return c.RenderJson(rs)
	}
}

func (c *Controller) renderJsonSucceed(data interface{}) (err error) {
	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	if jsoncallback != "" {
		return c.RenderJQueryCallback(jsoncallback, NewSucceedResult(data))
	} else {
		return c.RenderJson(NewSucceedResult(data))
	}
}

func (c *Controller) renderXmlError(err interface{}) (e error) {
	//log.Fatal(err)
	rs := ConvertErrorResult(err)
	return c.RenderXml(rs)
}

func (c *Controller) renderXmlSucceed(data interface{}) (err error) {
	return c.RenderXml(NewSucceedResult(data))
}

func (c *Controller) ParseForm() (url.Values, error) {
	if c.Ctx.Form != nil {
		return c.Ctx.Form, nil
	}
	form, err := parseForm(c)
	if err != nil {
		return nil, err
	}
	c.Ctx.Form = form
	return form, nil
}

func parseForm(c *Controller) (url.Values, error) {
	contentType := c.Ctx.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := c.Ctx.Request.ParseMultipartForm(defaultMaxMemory)
		if err != nil {
			return nil, err
		}
	} else {
		err := c.Ctx.Request.ParseForm()
		if err != nil {
			return nil, err
		}
	}

	return c.Ctx.Request.Form, nil
}

const (
	maxValueLength   = 4096
	maxHeaderLines   = 1024
	chunkSize        = 4 << 10  // 4 KB chunks
	defaultMaxMemory = 32 << 20 // 32 MB
)
