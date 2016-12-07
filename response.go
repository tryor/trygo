package ssss

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
)

type response struct {
	http.ResponseWriter
	Ctx    *Context
	render *render
}

func newResponse(ctx *Context) *response {
	rw := &response{Ctx: ctx}
	rw.render = &render{rw: rw}
	return rw
}

func (this *response) Error(code int, message string) (err error) {
	this.ResponseWriter.WriteHeader(code)
	_, err = this.ResponseWriter.Write([]byte(message))
	return
}

func (this *response) ContentType(typ string) {
	ctype := getContentType(typ)
	if ctype != "" {
		this.ResponseWriter.Header().Set("Content-Type", ctype)
	} else {
		this.ResponseWriter.Header().Set("Content-Type", typ)
	}
}

func (this *response) AddHeader(hdr string, val interface{}) {
	if v, ok := val.(string); ok {
		this.ResponseWriter.Header().Add(hdr, v)
	} else {
		this.ResponseWriter.Header().Add(hdr, fmt.Sprint(val))
	}
}

func (this *response) SetHeader(hdr string, val interface{}) {
	if v, ok := val.(string); ok {
		this.ResponseWriter.Header().Set(hdr, v)
	} else {
		this.ResponseWriter.Header().Set(hdr, fmt.Sprint(val))
	}
}

func (this *response) Flush() {
	if f, ok := this.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (this *response) CloseNotify() <-chan bool {
	if cn, ok := this.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	return nil
}

func (this *response) AddCookie(c *http.Cookie) {
	this.Header().Add("Set-Cookie", c.String())
}

type render struct {
	rw *response

	format       string
	contentType  string
	jsoncallback string
	//layout       bool
	wrap     bool
	wrapCode int //包装的消息code
	noWrap   bool
	gzip     bool //暂时未实现

	//数据
	status int //http status

	data interface{}
	err  error

	prepareDataFunc func()

	//标记是否已经开始
	started bool

	//标记是否已经被取消渲染
	canceled int32
}

func (this *render) String() string {
	length := -1
	if d, ok := this.data.([]byte); ok {
		length = len(d)
	}
	return fmt.Sprintf("Render: started:%v, format:%s, contentType:%s, jsoncallback:%s, wrap:%v, status:%d, len(data):%d, error:%v", this.started, this.format, this.contentType, this.jsoncallback, this.wrap, this.status, length, this.err)
}

func (this *render) Cancel() {
	atomic.StoreInt32(&this.canceled, 1)
}

func (this *render) IsCanceled(clear ...bool) bool {
	if len(clear) > 0 && clear[0] {
		return atomic.SwapInt32(&this.canceled, 0) > 0
	} else {
		return atomic.LoadInt32(&this.canceled) > 0
	}
}

func (this *render) Status(c int) *render {
	this.status = c
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

func (this *render) Wrap(code ...int) *render {
	this.wrap = true
	if len(code) > 0 {
		this.wrapCode = code[0]
	}
	return this
}

func (this *render) Nowrap(b ...bool) *render {
	this.noWrap = true
	return this
}

func (this *render) Gzip() *render {
	this.gzip = true
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
		this.jsoncallback = jsoncallback
	}
	if this.format == "" {
		this.Json()
	}
	return this
}

func (this *render) Xml() *render {
	this.format = FORMAT_XML
	this.contentType = "xml"
	return this
}

func (this *render) Header(key string, value ...interface{}) *render {
	h := this.rw.Header()
	if len(value) == 0 {
		h.Set(key, "")
		return this
	}
	if len(value) == 1 {
		h.Set(key, toString(value[0]))
		return this
	}
	for _, v := range value {
		h.Add(key, toString(v))
	}
	return this
}

func (this *render) Cookie(c *http.Cookie) *render {
	this.rw.AddCookie(c)
	return this
}

func (this *render) File(filename string) *render {
	if this.prepareDataFunc != nil {
		this.Reset()
		panic("Render: data already exists")
	}
	this.prepareDataFunc = func() {
		if this.contentType == "" {
			if idx := strings.LastIndex(filename, "."); idx != -1 {
				this.contentType = filename[idx:]
			}
		}
		this.data, this.err = os.Open(filename)
		if this.err != nil {
			Logger.Error("open file error:%v, filename:%s", this.err, filename)
		}
	}
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
			//如果设置了wrap, 将默认为json格式
			this.Json()
		}

		if this.status >= 400 || isErrorResult(data) || this.wrapCode != ERROR_CODE_OK {
			if this.jsoncallback == "" {
				this.data, this.err = BuildError(data, this.wrap, this.wrapCode, this.format)
			} else {
				this.data, this.err = BuildError(data, this.wrap, this.wrapCode, this.format, this.jsoncallback)
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
	return this
}

func (this *render) Reset() {
	this.contentType = ""
	this.data = nil
	this.format = ""
	this.jsoncallback = ""
	this.prepareDataFunc = nil
	this.err = nil
	this.status = 0
	this.started = false
	this.wrap = false
	this.noWrap = false
	this.wrapCode = 0
	this.gzip = false
	//this.canceled = 0
}

func (this *render) Exec() error {
	defer this.Reset()

	if !this.started {
		return errors.New("the render is not started")
	}

	if this.IsCanceled(true) {
		return nil
	}

	cfg := this.rw.Ctx.App.Config

	if this.noWrap {
		if this.wrap {
			this.wrap = false
		}
	} else {
		if !this.wrap && cfg.Render.Wrap {
			this.wrap = cfg.Render.Wrap
		}
	}

	if cfg.Render.AutoParseFormat && this.format == "" {
		this.format = this.rw.Ctx.Input.GetValue(cfg.Render.FormatParamName)
	}

	if cfg.Render.AutoParseFormat && this.jsoncallback == "" {
		this.jsoncallback = this.rw.Ctx.Input.GetValue(cfg.Render.JsoncallbackParamName)
	}

	//Logger.Debug("this.format:%v", this.format)
	//Logger.Debug("this.wrap:%v", this.wrap)

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
		return renderError(this.rw, this.err, http.StatusInternalServerError, this.wrap, ERROR_CODE_RUNTIME, this.format, this.jsoncallback)
	}

	var encoding string
	if cfg.Render.Gzip || this.gzip {
		encoding = ParseEncoding(this.rw.Ctx.Request)
	}

	this.rw.Header().Set("Content-Type", getContentType(this.contentType))
	switch data := this.data.(type) {
	case []byte:
		if _, _, err := WriteBody(encoding, this.rw, data, func(encodingEnable bool, name string) error {
			if encodingEnable {
				this.rw.SetHeader("Content-Encoding", name)
			} else {
				this.rw.SetHeader("Content-Length", strconv.Itoa(len(data)))
			}
			if this.status > 0 {
				this.rw.WriteHeader(this.status)
			}
			return nil
		}); err != nil {
			Logger.Warn("write data error, %v", err)
			//this.err = err
		}
	case *os.File:
		defer data.Close()
		if _, _, err := WriteFile(encoding, this.rw, data, func(encodingEnable bool, name string) error {
			if encodingEnable {
				this.rw.SetHeader("Content-Encoding", name)
			} else {
				stat, err := data.Stat()
				if err != nil {
					Logger.Error("stat file size error, %v", err)
					this.err = err
					return err
				} else {
					this.rw.SetHeader("Content-Length", strconv.FormatInt(stat.Size(), 10))
				}
			}
			if this.status > 0 {
				this.rw.WriteHeader(this.status)
			}
			return nil
		}); err != nil {
			Logger.Warn("write file error, %v", err)
			//this.err = err
		}
	default:
		this.err = errors.New("data type not supported")
		Logger.Error("%v", this.err)
	}

	if this.err != nil && !this.IsCanceled(true) {
		return renderError(this.rw, this.err, http.StatusInternalServerError, this.wrap, ERROR_CODE_RUNTIME, this.format, this.jsoncallback)
	}
	return this.err
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

func RenderFile(ctx *Context, filename string) *render {
	return Render(ctx).File(filename)
}

func RenderTemplate(ctx *Context, templateName string, data map[interface{}]interface{}) *render {
	return Render(ctx).Template(templateName, data)
}

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
func renderError(rw http.ResponseWriter, errdata interface{}, status int, wrap bool, wrapcode int, fmtAndJsoncallback ...string) error {
	var format, jsoncallback string
	if len(fmtAndJsoncallback) > 0 {
		format = fmtAndJsoncallback[0]
	} else if len(fmtAndJsoncallback) > 1 {
		jsoncallback = fmtAndJsoncallback[1]
	}

	var content []byte
	var err error
	if jsoncallback == "" {
		content, err = BuildError(errdata, wrap, wrapcode, format)
	} else {
		content, err = BuildError(errdata, wrap, wrapcode, format, jsoncallback)
	}

	if err != nil {
		//http.Error(rw, err.Error(), http.StatusInternalServerError)
		Logger.Error("format:%v, error:%v, data:%v", format, err, errdata)
		return err
	}
	return renderData(rw, toContentType(format), content, status)
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

func renderBuffer(rw http.ResponseWriter, contentType string, buff *bytes.Buffer, status ...int) error {
	rw.Header().Set("Content-Type", getContentType(contentType))
	if len(status) > 0 && status[0] != 0 {
		rw.WriteHeader(status[0])
	}
	_, err := io.Copy(rw, buff)
	if err != nil {
		Logger.Error("error:%v, buff.length:%v", err, buff.Len())
		return err
	}
	return nil
}

func renderData(rw http.ResponseWriter, contentType string, data []byte, status ...int) error {
	rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
	rw.Header().Set("Content-Type", getContentType(contentType))
	if len(status) > 0 && status[0] != 0 {
		rw.WriteHeader(status[0])
	}
	_, err := rw.Write(data)
	if err != nil {
		Logger.Error("error:%v, data.length:%v", err, len(data))
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

type root struct {
	Data interface{} `xml:"data"`
}

func buildXml(data interface{}) ([]byte, error) {
	switch reflect.TypeOf(data).Kind() {
	case reflect.String:
		return []byte(data.(string)), nil
	case reflect.Slice:
		if content, ok := data.([]byte); ok {
			return content, nil
		}
		//如果是reflect.Slice类型，需要将其放到一个根节点中
		data = root{Data: data}
	}

	content, err := xml.Marshal(data)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func buildJson(data interface{}) ([]byte, error) {
	switch jdata := data.(type) {
	case []byte:
		//如果是[]byte类型，就认为已经是标准json格式数据
		return jdata, nil
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
	case []byte:
		//如果是[]byte类型，就认为已经是标准json格式数据
		content.Write(data.([]byte))
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
func BuildError(err interface{}, wrap bool, code int, format string, jsoncallback ...string) ([]byte, error) {
	switch format {
	case "json":
		return buildJsonError(err, wrap, code, jsoncallback...)
	case "xml":
		return buildXmlError(err, wrap, code)
	default:
		return buildData(err), nil
	}
}

func buildJsonError(err interface{}, wrap bool, code int, jsoncallback ...string) ([]byte, error) {
	if wrap {
		err = convertErrorResult(err, code)
	}
	if len(jsoncallback) > 0 && jsoncallback[0] != "" {
		return buildJQueryCallback(jsoncallback[0], err)
	} else {
		return buildJson(err)
	}
}

func buildXmlError(err interface{}, wrap bool, code int) ([]byte, error) {
	if wrap {
		err = convertErrorResult(err, code)
	}
	return buildXml(err)
}
