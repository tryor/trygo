package ssss

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	//log "github.com/cihub/seelog"
	"io"
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
	Prepare()
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
}

func (c *Controller) Init(app *App, ctx *Context, cn string, mn string) {
	c.Data = make(map[interface{}]interface{})
	c.App = app
	c.ChildName = cn
	c.MethodName = mn
	c.Ctx = ctx
	c.TplNames = ""
	c.TplExt = "tpl"
}

func (c *Controller) Prepare() {

}

func (c *Controller) Finish() {

}

func (c *Controller) Redirect(url string, code int) {
	c.Ctx.Redirect(code, url)
}

func (c *Controller) Error(code int, message string) {
	c.Ctx.Error(code, message)
}

func (c *Controller) Render(contentType string, data []byte) {
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(data)))
	c.Ctx.ContentType(contentType)
	c.Ctx.ResponseWriter.Write(data)
}

func (c *Controller) RenderHtml(content string) {
	c.Render("html", []byte(content))
}

func (c *Controller) RenderText(content string) {
	c.Render("txt", []byte(content))
}

func (c *Controller) RenderJson(data interface{}) {
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Render("application/json", content)
}

func (c *Controller) RenderJQueryCallback(jsoncallback string, data interface{}) {
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
	c.Render("application/json", bjson)
}

func (c *Controller) RenderXml(data interface{}) {
	content, err := xml.Marshal(data)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Render("xml", content)
}

func (c *Controller) RenderTemplate(contentType ...string) {
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
	err := t.ExecuteTemplate(ibytes, file, c.Data)
	if err != nil {
		log.Println("template Execute err:", err)
	}
	icontent, _ := ioutil.ReadAll(ibytes)
	if len(contentType) > 0 {
		c.Render(contentType[0], icontent)
	} else {
		c.Render("html", icontent)
	}
}

//fmt值指示响应结果格式，当前支持:json或xml, 默认为:json
//如果是json格式结果，支持jsoncallback
func (c *Controller) RenderError(fmt string, err interface{}) {
	//fmt := c.Ctx.Request.FormValue("fmt")
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
		c.renderJsonError(err)
	case "xml":
		c.renderXmlError(err)
	default:
		c.renderJsonError(err)
	}
}

//fmt值指示响应结果格式，当前支持:json或xml, 默认为:json
//如果是json格式结果，支持jsoncallback
func (c *Controller) RenderSucceed(fmt string, data interface{}) {
	//fmt := c.Ctx.Request.FormValue("fmt")
	fmt = strings.ToLower(fmt)
	switch fmt {
	case "":
		fallthrough
	case "json":
		c.renderJsonSucceed(data)
	case "xml":
		c.renderXmlSucceed(data)
	default:
		c.renderJsonSucceed(data)
	}
}

func (c *Controller) renderJsonError(err interface{}) {
	//log.Fatal(err)
	rs := convertErrorResult(err)

	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	if jsoncallback != "" {
		c.RenderJQueryCallback(jsoncallback, rs)
	} else {
		c.RenderJson(rs)
	}
}

func (c *Controller) renderJsonSucceed(data interface{}) {
	jsoncallback := c.Ctx.Request.FormValue("jsoncallback")
	if jsoncallback != "" {
		c.RenderJQueryCallback(jsoncallback, NewSucceedResult(data))
	} else {
		c.RenderJson(NewSucceedResult(data))
	}
}

func (c *Controller) renderXmlError(err interface{}) {
	//log.Fatal(err)
	rs := convertErrorResult(err)
	c.RenderXml(rs)
}

func (c *Controller) renderXmlSucceed(data interface{}) {
	c.RenderXml(NewSucceedResult(data))
}

func (c *Controller) ParseForm() (url.Values, error) {
	//defer func() {
	//	if c.Ctx.Request.Body != nil {
	//		c.Ctx.Request.Body.Close()
	//	}
	//}()

	var form url.Values
	contentType := c.Ctx.Request.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" && c.Ctx.Request.ContentLength > 0 && c.Ctx.Request.Body != nil {
		body := make([]byte, c.Ctx.Request.ContentLength)
		_, err := io.ReadFull(c.Ctx.Request.Body, body)
		if err != nil {
			return nil, err
		}
		bodystr := string(body)
		form, err = url.ParseQuery(bodystr)
		if err != nil {
			return nil, err
		}
	} else {
		err := c.Ctx.Request.ParseForm()
		if err != nil {
			return nil, err
		}
		form = c.Ctx.Request.Form
		//log.Info("access:", form)
	}
	return form, nil
}
