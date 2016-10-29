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
	if err != nil {
		log.Printf("error:%v, data:%v\n", err, data)
	}
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
		log.Printf("error:%v, data:%v\n", err, data)
		return err
	}
	return c.Render("application/json", content)
}

func (c *Controller) RenderJQueryCallback(jsoncallback string, data interface{}) error {
	bjson, err := buildJQueryCallback(jsoncallback, data)
	if err != nil {
		log.Printf("error:%v, data:%v\n", err, data)
	}
	return c.Render("application/json", bjson)
}

func (c *Controller) RenderXml(data interface{}) error {
	content, err := xml.Marshal(data)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		log.Printf("error:%v, data:%v\n", err, data)
		return err
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

	if c.App.Config.RunMode == RUNMODE_DEV {
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

func (c *Controller) RenderData(fmt string, data []byte) error {
	switch fmt {
	case "":
		fallthrough
	case "json":
		return c.Render("application/json", data)
	case "xml":
		return c.Render("xml", data)
	default:
		return c.Render("application/json", data)
	}
}

//fmt值指示响应结果格式，当前支持:json或xml, 默认为:json
//如果是json格式结果，支持jsoncallback
<<<<<<< HEAD
func (c *Controller) RenderError(fmt string, errdata interface{}) error {
	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	//fmt = strings.ToLower(fmt)
	var content []byte
	var err error
	if jsoncallback == "" {
		content, err = BuildError(errdata, fmt)
	} else {
		content, err = BuildError(errdata, fmt, jsoncallback)
	}

=======
func (c *Controller) RenderError(fmt string, errdata interface{}) (e error) {
	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	//fmt = strings.ToLower(fmt)
	content, err := BuildError(errdata, fmt, jsoncallback)
>>>>>>> d913fd4e243a2a6f59a580b381ead680a510c335
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		log.Printf("format:%v, error:%v, data:%v\n", fmt, err, errdata)
		return err
	}
	return c.RenderData(fmt, content)
}

//fmt值指示响应结果格式，当前支持:json或xml, 默认为:json
//如果是json格式结果，支持jsoncallback
<<<<<<< HEAD
func (c *Controller) RenderSucceed(fmt string, data interface{}) error {
	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	//fmt = strings.ToLower(fmt)
	var content []byte
	var err error
	if jsoncallback == "" {
		content, err = BuildSucceed(data, fmt)
	} else {
		content, err = BuildSucceed(data, fmt, jsoncallback)
	}
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		log.Printf("format:%v, error:%v, data:%v\n", fmt, err, data)
		return err
	}
=======
func (c *Controller) RenderSucceed(fmt string, data interface{}) (err error) {
	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	//fmt = strings.ToLower(fmt)
	content, err := BuildSucceed(data, fmt, jsoncallback)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		log.Printf("format:%v, error:%v, data:%v\n", fmt, err, data)
		return err
	}
>>>>>>> d913fd4e243a2a6f59a580b381ead680a510c335
	return c.RenderData(fmt, content)
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

//fmt 结果格式, 值有：json, xml
//jsoncallback 当需要将json结果做为js函数参数时，在jsoncallback中指定函数名
func BuildSucceed(data interface{}, fmt string, jsoncallback ...string) ([]byte, error) {
	//fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
		return buildJsonSucceed(data, jsoncallback...)
	case "xml":
		return buildXmlSucceed(data)
	default:
		return buildJsonSucceed(data, jsoncallback...)
	}
}

func buildJsonSucceed(data interface{}, jsoncallback ...string) ([]byte, error) {
<<<<<<< HEAD
	if len(jsoncallback) > 0 && jsoncallback[0] != "" {
=======
	if len(jsoncallback) > 0 {
>>>>>>> d913fd4e243a2a6f59a580b381ead680a510c335
		return buildJQueryCallback(jsoncallback[0], NewSucceedResult(data))
	} else {
		return buildJson(NewSucceedResult(data))
	}
}

func buildXmlSucceed(data interface{}) ([]byte, error) {
	return buildXml(NewSucceedResult(data))
}

func buildXml(data interface{}) ([]byte, error) {
	content, err := xml.Marshal(data)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func buildJson(data interface{}) ([]byte, error) {
	content, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func buildJQueryCallback(jsoncallback string, data interface{}) ([]byte, error) {
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
			return nil, err
		}
	}

	bjson := []byte(jsoncallback)
	bjson = append(bjson, '(')
	bjson = append(bjson, content...)
	bjson = append(bjson, ')')
	return bjson, nil
}

//fmt 结果格式, 值有：json, xml
//jsoncallback 当需要将json结果做为js函数参数时，在jsoncallback中指定函数名
func BuildError(err interface{}, fmt string, jsoncallback ...string) ([]byte, error) {
	//fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
<<<<<<< HEAD
		return buildJsonError(err, jsoncallback...)
	case "xml":
		return buildXmlError(err)
	default:
		return buildJsonError(err, jsoncallback...)
=======
		return buildJsonError(err)
	case "xml":
		return buildXmlError(err)
	default:
		return buildJsonError(err)
>>>>>>> d913fd4e243a2a6f59a580b381ead680a510c335
	}
}

func buildJsonError(err interface{}, jsoncallback ...string) ([]byte, error) {
	rs := ConvertErrorResult(err)
<<<<<<< HEAD
	if len(jsoncallback) > 0 && jsoncallback[0] != "" {
=======
	if len(jsoncallback) > 0 {
>>>>>>> d913fd4e243a2a6f59a580b381ead680a510c335
		return buildJQueryCallback(jsoncallback[0], rs)
	} else {
		return buildJson(rs)
	}
}

func buildXmlError(err interface{}) ([]byte, error) {
	rs := ConvertErrorResult(err)
	return buildXml(rs)
}
