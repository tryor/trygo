package ssss

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
)

type Response struct {
	http.ResponseWriter
	Ctx    *Context
	render *render
}

func (this *Response) Error(code int, message string) (err error) {
	this.ResponseWriter.WriteHeader(code)
	_, err = this.ResponseWriter.Write([]byte(message))
	return
}

func (this *Response) ContentType(typ string) {
	ctype := getContentType(typ)
	if ctype != "" {
		this.ResponseWriter.Header().Set("Content-Type", ctype)
	} else {
		this.ResponseWriter.Header().Set("Content-Type", typ)
	}
}

func (this *Response) AddHeader(hdr string, val string) {
	this.ResponseWriter.Header().Add(hdr, val)
}

func (this *Response) SetHeader(hdr string, val string) {
	this.ResponseWriter.Header().Set(hdr, val)
}

func (this *Response) Flush() {
	if f, ok := this.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (this *Response) CloseNotify() <-chan bool {
	if cn, ok := this.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	return nil
}

type render struct {
	rw *Response

	format       string
	contentType  string
	jsoncallback string
	//layout       bool
	wrap bool
	gzip bool //暂时未实现

	//数据
	code int
	data []byte
	err  error

	prepareDataFunc func()

	//标记是否已经开始
	started bool
}

func (this *render) String() string {
	return fmt.Sprintf("Render: started:%v, format:%s, contentType:%s, jsoncallback:%s, wrap:%v, code:%d, len(data):%d, error:%v", this.started, this.format, this.contentType, this.jsoncallback, this.wrap, this.code, len(this.data), this.err)
}

func (this *render) Code(c int) *render {
	this.code = c
	return this
}

func (this *render) ContentType(typ string) *render {
	this.contentType = typ
	return this
}

//结果格式, json or xml or txt or html or other
func (this *render) Format(format string) *render {
	this.format = format
	return this
}

func (this *render) Wrap() *render {
	this.wrap = true
	return this
}

func (this *render) Html() *render {
	this.format = FORMAT_HTML
	this.contentType = "html"
	return this
}

func (this *render) Text() *render {
	this.format = FORMAT_TXT
	this.contentType = "txt"
	return this
}

func (this *render) Json() *render {
	this.format = FORMAT_JSON
	this.contentType = "application/json; charset=utf-8"
	return this
}

func (this *render) JsonCallback(jsoncallback string) *render {
	if jsoncallback != "" {
		if this.format == "" {
			this.Json()
		}
		this.jsoncallback = jsoncallback
	}
	return this
}

func (this *render) Xml() *render {
	this.format = FORMAT_XML
	this.contentType = "xml"
	return this
}

func (this *render) Template(templateName string, data map[interface{}]interface{}) *render {
	if this.prepareDataFunc != nil {
		this.Reset()
		panic("Render: data already exists")
	}
	this.prepareDataFunc = func() {
		if this.contentType == "" {
			this.Html()
		}
		this.data, this.err = BuildTemplateData(this.rw.Ctx.App, templateName, data)
		if this.err != nil {
			Logger.Error("template execute error:%v, template:%s", this.err, templateName)
		}
	}
	return this
}

//data - 如果data为[]byte类型，将直接输出，不再会进行json,xml等编码
func (this *render) Data(data interface{}) *render {
	if this.prepareDataFunc != nil {
		this.Reset()
		panic("Render: data already exists")
	}

	this.prepareDataFunc = func() {

		if this.wrap && this.format == "" {
			this.Json()
		}

		if bs, ok := data.([]byte); ok {
			this.data = bs
		} else {
			if this.code >= 400 {
				if this.jsoncallback == "" {
					this.data, this.err = BuildError(data, this.wrap, this.format)
				} else {
					this.data, this.err = BuildError(data, this.wrap, this.format, this.jsoncallback)
				}
			} else {
				if this.jsoncallback == "" {
					this.data, this.err = BuildSucceed(data, this.wrap, this.format)
				} else {
					this.data, this.err = BuildSucceed(data, this.wrap, this.format, this.jsoncallback)
				}
			}

			if this.err != nil {
				Logger.Error("error:%v, data:%v", this.err, data)
			}
		}

		//		if this.contentType == "" {
		//			if _, ok := data.(string); ok {
		//				this.Text()
		//			}
		//		}

	}
	return this
}

func (this *render) Reset() {
	this.contentType = ""
	this.data = nil
	this.format = ""
	this.jsoncallback = ""
	this.prepareDataFunc = nil
	this.err = nil
	this.code = 0
	this.started = false
	this.wrap = this.rw.Ctx.App.Config.RenderWrap
}

func (this *render) Exec() error {
	defer this.Reset()

	if !this.started {
		return errors.New("the render is not started")
	}

	cfg := this.rw.Ctx.App.Config

	if !this.wrap && cfg.RenderWrap {
		this.wrap = cfg.RenderWrap
	}

	if cfg.AutoParseRenderFormat && this.format == "" {
		this.format = this.rw.Ctx.Request.FormValue(cfg.FormatParamName)
	}

	if cfg.AutoParseRenderFormat && this.jsoncallback == "" {
		this.jsoncallback = this.rw.Ctx.Request.FormValue(cfg.JsoncallbackParamName)
	}

	if this.prepareDataFunc != nil {
		this.prepareDataFunc()
	}

	if this.contentType == "" {
		this.contentType = toContentType(this.format)
	}
	if this.contentType == "" {
		this.contentType = "txt"
	}

	if this.err != nil {
		return renderError(this.rw, this.err, http.StatusInternalServerError, this.wrap, this.format, this.jsoncallback)
	}

	return renderData(this.rw, this.contentType, this.data, this.code)
}

func Render(ctx *Context, data ...interface{}) *render {
	render := ctx.ResponseWriter.render
	if render.started {
		panic("Render: is already started")
	}

	render.started = true
	if len(data) > 0 {
		if len(data) == 1 {
			render.Data(data[0])
		} else {
			render.Data(data)
		}
	}
	return render
}

func RenderTemplate(ctx *Context, templateName string, data map[interface{}]interface{}) *render {
	return Render(ctx).Template(templateName, data)
}

//func RenderHtml(rw http.ResponseWriter, content string) (err error) {
//	return Render(rw, "html", []byte(content))
//}

//func RenderText(rw http.ResponseWriter, content string) (err error) {
//	return Render(rw, "txt", []byte(content))
//}

//func RenderJson(rw http.ResponseWriter, data interface{}) (err error) {
//	content, err := buildJson(data)
//	if err != nil {
//		http.Error(rw, err.Error(), http.StatusInternalServerError)
//		Logger.Error("error:%v, data:%v", err, data)
//		return err
//	}
//	return Render(rw, "application/json", content)
//}

//func RenderJQueryCallback(rw http.ResponseWriter, jsoncallback string, data interface{}) error {
//	bjson, err := buildJQueryCallback(jsoncallback, data)
//	if err != nil {
//		http.Error(rw, err.Error(), http.StatusInternalServerError)
//		Logger.Error("error:%v, data:%v", err, data)
//		return err
//	}
//	return Render(rw, "application/json", bjson)
//}

//func RenderXml(rw http.ResponseWriter, data interface{}) error {
//	content, err := buildXml(data)
//	if err != nil {
//		http.Error(rw, err.Error(), http.StatusInternalServerError)
//		Logger.Error("error:%v, data:%v", err, data)
//		return err
//	}
//	return Render(rw, "xml", content)
//}

//func RenderTemplate(rw http.ResponseWriter, app *App, tplnames string, data map[interface{}]interface{}, contentType ...string) (err error) {
//	content, err := buildTemplateData(app, tplnames, data)
//	if err != nil {
//		http.Error(rw, err.Error(), http.StatusInternalServerError)
//		Logger.Error("template execute error:%v, tplnames:%s", err, tplnames)
//		return err
//	}
//	if len(contentType) > 0 {
//		return Render(rw, contentType[0], content)
//	} else {
//		return Render(rw, "html", content)
//	}
//}

func BuildTemplateData(app *App, tplnames string, data map[interface{}]interface{}) ([]byte, error) {

	_, file := path.Split(tplnames)
	subdir := path.Dir(tplnames)
	ibytes := bytes.NewBufferString("")

	if app.Config.RunMode == DEV {
		app.buildTemplate()
	}

	t := app.TemplateRegister.Templates[subdir]
	if t == nil {
		return nil, errors.New(fmt.Sprintf("template not exist, tplnames:%s", tplnames))
	}
	err := t.ExecuteTemplate(ibytes, file, data)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(ibytes)
	if err != nil {
		return nil, err
	}
	return content, nil
}

//fmtAndJsoncallback[0] - format, 值指示响应结果格式，当前支持:json或xml, 默认为:json
//fmtAndJsoncallback[1] - jsoncallback 如果是json格式结果，支持jsoncallback
func renderError(rw http.ResponseWriter, errdata interface{}, code int, wrap bool, fmtAndJsoncallback ...string) error {
	var format, jsoncallback string
	if len(fmtAndJsoncallback) > 0 {
		format = fmtAndJsoncallback[0]
	} else if len(fmtAndJsoncallback) > 1 {
		jsoncallback = fmtAndJsoncallback[1]
	}

	var content []byte
	var err error
	if jsoncallback == "" {
		content, err = BuildError(errdata, wrap, format)
	} else {
		content, err = BuildError(errdata, wrap, format, jsoncallback)
	}

	if err != nil {
		//http.Error(rw, err.Error(), http.StatusInternalServerError)
		Logger.Error("format:%v, error:%v, data:%v", format, err, errdata)
		return err
	}
	return renderData(rw, toContentType(format), content, code)
}

//fmtAndJsoncallback[0] - fmt, 值指示响应结果格式，当前支持:json或xml, 默认为:json
//fmtAndJsoncallback[1] - jsoncallback 如果是json格式结果，支持jsoncallback
func renderSucceed(rw http.ResponseWriter, data interface{}, wrap bool, fmtAndJsoncallback ...string) error {
	var format, jsoncallback string
	if len(fmtAndJsoncallback) > 0 {
		format = fmtAndJsoncallback[0]
	} else if len(fmtAndJsoncallback) > 1 {
		jsoncallback = fmtAndJsoncallback[1]
	}
	var content []byte
	var err error
	if jsoncallback == "" {
		content, err = BuildSucceed(data, wrap, format)
	} else {
		content, err = BuildSucceed(data, wrap, format, jsoncallback)
	}
	if err != nil {
		//http.Error(rw, err.Error(), http.StatusInternalServerError)
		Logger.Error("format:%v, error:%v, data:%v", format, err, data)
		return err
	}
	return renderData(rw, toContentType(format), content)
}

//func renderData(rw http.ResponseWriter, format string, data []byte, code ...int) error {
//	return Render(rw, toContentType(format), data, code...)
//}

func renderData(rw http.ResponseWriter, contentType string, data []byte, code ...int) error {
	rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
	rw.Header().Set("Content-Type", getContentType(contentType))
	if len(code) > 0 && code[0] != 0 {
		rw.WriteHeader(code[0])
	}
	_, err := rw.Write(data)
	if err != nil {
		Logger.Error("error:%v, data:%v", err, data)
		return err
	}
	return nil
}

//format 结果格式, 值有：json, xml, 其它(txt, html, ...)
//jsoncallback 当需要将json结果做为js函数参数时，在jsoncallback中指定函数名
func BuildSucceed(data interface{}, wrap bool, format string, jsoncallback ...string) ([]byte, error) {
	switch format {
	case "json":
		return buildJsonSucceed(data, wrap, jsoncallback...)
	case "xml":
		return buildXmlSucceed(data, wrap)
	default:
		return buildData(data), nil
	}
}

func buildData(data interface{}) []byte {
	switch d := data.(type) {
	case string:
		return []byte(d)
	case []byte:
		return d
	default:
		return []byte(fmt.Sprint(data))
	}
}

func buildJsonSucceed(data interface{}, wrap bool, jsoncallback ...string) ([]byte, error) {
	if wrap {
		data = NewSucceedResult(data)
	}
	if len(jsoncallback) > 0 && jsoncallback[0] != "" {
		return buildJQueryCallback(jsoncallback[0], data)
	} else {
		return buildJson(data)
	}
}

func buildXmlSucceed(data interface{}, wrap bool) ([]byte, error) {
	if wrap {
		data = NewSucceedResult(data)
	}
	return buildXml(data)
}

func buildXml(data interface{}) ([]byte, error) {
	if s, ok := data.(string); ok {
		return []byte(s), nil
	}
	content, err := xml.Marshal(data)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func buildJson(data interface{}) ([]byte, error) {
	switch jdata := data.(type) {
	case string:
		content := bytes.NewBuffer(make([]byte, 0, len(jdata)))
		content.WriteByte('"')
		content.WriteString(jdata)
		content.WriteByte('"')
		return content.Bytes(), nil
	case []byte:
		content := bytes.NewBuffer(make([]byte, 0, len(jdata)))
		content.WriteByte('"')
		content.Write(jdata)
		content.WriteByte('"')
		return content.Bytes(), nil
	default:
		jsondata, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return jsondata, nil
	}

}

func buildJQueryCallback(jsoncallback string, data interface{}) ([]byte, error) {
	content := bytes.NewBuffer(make([]byte, 0))
	content.WriteString(jsoncallback)
	content.WriteByte('(')
	switch data.(type) {
	case string:
		content.WriteByte('"')
		content.WriteString(data.(string))
		content.WriteByte('"')
	case []byte:
		content.WriteByte('"')
		content.Write(data.([]byte))
		content.WriteByte('"')
	default:
		jsondata, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		content.Write(jsondata)
	}

	content.WriteByte(')')
	return content.Bytes(), nil
}

//format 结果格式, 值有：json, xml, 其它(txt, html, ...)
//jsoncallback 当需要将json结果做为js函数参数时，在jsoncallback中指定函数名
func BuildError(err interface{}, wrap bool, format string, jsoncallback ...string) ([]byte, error) {
	switch format {
	case "json":
		return buildJsonError(err, wrap, jsoncallback...)
	case "xml":
		return buildXmlError(err, wrap)
	default:
		return buildData(err), nil
	}
}

func buildJsonError(err interface{}, wrap bool, jsoncallback ...string) ([]byte, error) {
	if wrap {
		err = convertErrorResult(err)
	}
	if len(jsoncallback) > 0 && jsoncallback[0] != "" {
		return buildJQueryCallback(jsoncallback[0], err)
	} else {
		return buildJson(err)
	}
}

func buildXmlError(err interface{}, wrap bool) ([]byte, error) {
	if wrap {
		err = convertErrorResult(err)
	}
	return buildXml(err)
}
