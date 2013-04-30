package ssss

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
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

func (c *Controller) Render(contentType string, data []byte) {
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(data)), true)
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
	c.Render("json", content)
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
	c.Render("json", bjson)
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
