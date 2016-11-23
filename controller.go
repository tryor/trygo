package ssss

import (
	"net/http"
)

//	"net/url"

type IController interface {
	Init(app *App, ctx *Context, controllerName, methodName string)
	Prepare()
	Get()
	Post()
	Delete()
	Put()
	Head()
	Patch()
	Options()
	Finish()
}

type Controller struct {
	Data           map[interface{}]interface{}
	Ctx            *Context
	ControllerName string
	MethodName     string
	TplNames       string
	TplExt         string
	App            *App
}

func (c *Controller) Init(app *App, ctx *Context, controllerName, methodName string) {
	c.Data = make(map[interface{}]interface{})
	c.App = app
	c.ControllerName = controllerName
	c.MethodName = methodName
	c.Ctx = ctx
	c.TplExt = "tpl"
}

func (c *Controller) Prepare() {
}

func (c *Controller) Finish() {
}

func (c *Controller) Get() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (c *Controller) Post() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (c *Controller) Delete() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (c *Controller) Put() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (c *Controller) Head() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (c *Controller) Patch() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (c *Controller) Options() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (c *Controller) Redirect(url string, code int) {
	c.Ctx.Redirect(code, url)
}

func (c *Controller) Error(code int, message string) {
	c.Ctx.ResponseWriter.Error(code, message)
}

func (c *Controller) Render(data ...interface{}) *render {
	return Render(c.Ctx, data...)
}

func (c *Controller) RenderTemplate() *render {
	if c.TplNames == "" {
		c.TplNames = c.ControllerName + "/" + c.MethodName + "." + c.TplExt
	}
	return RenderTemplate(c.Ctx, c.TplNames, c.Data)
}

//func (c *Controller) Render(contentType string, data []byte) (err error) {
//	return Render(c.Ctx.ResponseWriter, contentType, data)
//}

//func (c *Controller) RenderHtml(content string) (err error) {
//	return RenderHtml(c.Ctx.ResponseWriter, content)
//}

//func (c *Controller) RenderText(content string) (err error) {
//	return RenderText(c.Ctx.ResponseWriter, content)
//}

//func (c *Controller) RenderJson(data interface{}) (err error) {
//	return RenderJson(c.Ctx.ResponseWriter, data)
//}

//func (c *Controller) RenderJQueryCallback(jsoncallback string, data interface{}) error {
//	return RenderJQueryCallback(c.Ctx.ResponseWriter, jsoncallback, data)
//}

//func (c *Controller) RenderXml(data interface{}) error {
//	return RenderXml(c.Ctx.ResponseWriter, data)
//}

//func (c *Controller) RenderTemplate(contentType ...string) (err error) {
//	if c.TplNames == "" {
//		c.TplNames = c.ControllerName + "/" + c.MethodName + "." + c.TplExt
//	}
//	return RenderTemplate(c.Ctx.ResponseWriter, c.App, c.TplNames, c.Data, contentType...)
//}

//func (c *Controller) RenderData(format string, data []byte) error {
//	return RenderData(c.Ctx.ResponseWriter, format, data)
//}

//func (c *Controller) RenderError(err interface{}, code int, fmtAndJsoncallback ...string) error {
//	return RenderError(c.Ctx.ResponseWriter, err, code, fmtAndJsoncallback...)
//}

//func (c *Controller) RenderSucceed(data interface{}, fmtAndJsoncallback ...string) error {
//	return RenderSucceed(c.Ctx.ResponseWriter, data, fmtAndJsoncallback...)
//}
