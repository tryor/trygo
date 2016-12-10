package trygo

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type Context struct {
	App            *App
	ResponseWriter *response
	Request        *http.Request
	Multipart      bool
	Input          *input
	//某些错误会放在这里， 必要时可对此进行检查
	//比如，配置：Config.ThrowBindParamPanic = false, 且绑定参数发生错误时，可在此检查错误原因
	Error error
}

func newContext() *Context {
	ctx := &Context{}
	ctx.ResponseWriter = newResponse(ctx)
	ctx.Input = newInput(ctx)
	return ctx
}

func NewContext(rw http.ResponseWriter, r *http.Request, app *App) *Context {
	ctx := newContext()
	ctx.Reset(rw, r, app)
	return ctx
}

func (ctx *Context) Reset(rw http.ResponseWriter, r *http.Request, app *App) *Context {
	if resp, ok := rw.(*response); ok {
		ctx.ResponseWriter = resp
	} else {
		ctx.ResponseWriter.ResponseWriter = rw
	}
	ctx.Request = r
	ctx.App = app
	ctx.Multipart = strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data")
	ctx.Input.Values = nil
	ctx.ResponseWriter.render.Reset()
	return ctx
}

func (ctx *Context) Redirect(status int, url string) {
	http.Redirect(ctx.ResponseWriter, ctx.Request, url, status)
}

func (ctx *Context) Render(data ...interface{}) *render {
	return Render(ctx, data...)
}

func (ctx *Context) RenderTemplate(templateName string, data map[interface{}]interface{}) *render {
	return RenderTemplate(ctx, templateName, data)
}

func (ctx *Context) RenderFile(filename string) *render {
	return RenderFile(ctx, filename)
}

type input struct {
	url.Values
	ctx *Context
}

func newInput(ctx *Context) *input {
	inpt := &input{ctx: ctx}
	if ctx.Request != nil && ctx.Request.Form != nil {
		inpt.Values = ctx.Request.Form
	}
	return inpt
}

func (input *input) Parse() error {
	if input.Values != nil {
		return nil
	}

	if input.ctx.Request.Form != nil {
		input.Values = input.ctx.Request.Form
		return nil
	}

	form, err := parseForm(input.ctx.Request, input.ctx.Multipart)
	if err != nil {
		//Logger.Error("%v", err)
		return err
	}
	input.Values = form
	return nil
}

func (input *input) AddValues(values url.Values) {
	if input.Values == nil {
		input.Parse()
	}
	if input.Values != nil {
		for k, v := range values {
			input.Values[k] = append(input.Values[k], v...)
		}
	}
}

func (input *input) GetValue(key string) string {
	if input.Values == nil {
		err := input.Parse()
		if err != nil {
			return ""
		}
	}
	return input.Values.Get(key)
}

func (input *input) GetValues(key string) []string {
	if input.Values == nil {
		err := input.Parse()
		if err != nil {
			return []string{}
		}
	}
	return input.Values[key]
}

func getTaginfo(name string, taginfoses []Taginfos) *tagInfo {
	for _, tis := range taginfoses {
		if ti, ok := tis[name]; ok && ti != nil {
			return ti
		}
	}
	return nil
}

func (input *input) Bind(dest interface{}, key string, taginfos ...Taginfos) error {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		panic("non-pointer can not bind: " + key)
	}
	value = value.Elem()
	if !value.CanSet() {
		panic("non-settable variable can not bind: " + key)
	}

	typ := value.Type()
	isStruct := typ.Kind() == reflect.Struct
	if isStruct {
		taginfos = append(taginfos, parseTags(map[string]reflect.Type{key: typ}, []string{}, input.ctx.App.Config.FormDomainModel))
	}

	rv, err := input.bind(key, typ, taginfos...)
	if err != nil {
		input.ctx.Error = err
		if input.ctx.App.Config.ThrowBindParamPanic {
			//if input.ctx.App.Config.Render.Wrap {
			var msg string
			if isStruct {
				msg = fmt.Sprintf("%v, cause:%s.%v", ERROR_INFO_MAP[ERROR_CODE_PARAM_ILLEGAL], key, err)
			} else {
				msg = fmt.Sprintf("%v, %s=%v, cause:%v", ERROR_INFO_MAP[ERROR_CODE_PARAM_ILLEGAL], key, input.ctx.Input.Values[key], err)
			}
			panic(NewErrorResult(ERROR_CODE_PARAM_ILLEGAL, msg))
			//}
			//panic(err)
		}
		return err
	}
	if !rv.IsValid() {
		err := errors.New("reflect value not is valid")
		input.ctx.Error = err
		if input.ctx.App.Config.ThrowBindParamPanic {
			panic(err)
		}
		return err
	}
	value.Set(*rv)
	return nil
}

func (input *input) bind(pname string, ptype reflect.Type, taginfos ...Taginfos) (*reflect.Value, error) {
	ctx := input.ctx
	vp := reflect.Indirect(reflect.New(ptype))
	kind := ptype.Kind()
	switch kind {
	case reflect.Slice:
		kind = ptype.Elem().Kind()
		if reflect.Struct == kind {
			return nil, errors.New("the parameter slice type is not supported")
		} else {
			vals := ctx.Input.Values[pname]
			for _, str := range vals {
				v := reflect.Indirect(reflect.New(ptype.Elem()))
				if err := input.parseAndCheck(getTaginfo(pname, taginfos), kind, str, true, &v); err != nil {
					return nil, err
				}
				vp = reflect.Append(vp, v)
			}
		}
	case reflect.Struct:
		tp := vp.Type()
		for i := 0; i < tp.NumField(); i++ {
			f := tp.Field(i)
			var name string
			if f.Anonymous {
				name = pname
			} else {
				paramTag := strings.TrimSpace(f.Tag.Get("field"))
				if paramTag == "-" {
					continue
				}
				paramTags := strings.SplitN(paramTag, ",", 2)
				if len(paramTag) > 0 {
					name = strings.TrimSpace(paramTags[0])
					if name == "-" {
						continue
					}
				}
				if len(name) == 0 {
					name = strings.ToLower(f.Name[0:1]) + f.Name[1:]
				}
				//if _, ok := form[name]; !ok {
				//	name = f.Name
				//}
				if ctx.App.Config.FormDomainModel {
					name = pname + "." + name
				}
			}
			v, err := input.bind(name, f.Type, taginfos...)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%v=%v, %v", name, ctx.Input.Values[name], err))
			}
			vp.Field(i).Set(*v)
		}

	case reflect.Map:
		vp.Set(reflect.ValueOf(ctx.Input.Values))

	default:
		vals, ok := ctx.Input.Values[pname]
		var val string
		if ok {
			val = vals[0]
		}
		if err := input.parseAndCheck(getTaginfo(pname, taginfos), ptype.Kind(), val, ok, &vp); err != nil {
			return nil, err
		}
	}
	return &vp, nil
}

//在vp中返回值
func (input *input) parseAndCheck(tagInfo *tagInfo, kind reflect.Kind, val string, ok bool, vp *reflect.Value) error {
	if tagInfo != nil {
		return tagInfo.Check(val, vp)
	} else {
		_, err := parseValue(val, kind, vp)
		if err != nil {
			return err
		}
	}
	return nil
}

//func (ctx *Context) Abort(status int, body string) {
//	ctx.ResponseWriter.WriteHeader(status)
//	ctx.ResponseWriter.Write([]byte(body))
//}

//func (ctx *Context) NotModified() {
//	ctx.ResponseWriter.WriteHeader(304)
//}

//func (ctx *Context) NotFound(message string) {
//	ctx.Error(404, message)
//}

//func (ctx *Context) AddHeader(hdr string, val string) {
//	ctx.ResponseWriter.Header().Add(hdr, val)
//}

////func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
//func (ctx *Context) SetHeader(hdr string, val string) {
//	ctx.ResponseWriter.Header().Set(hdr, val)
//}

//func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
//	if unique {
//		ctx.ResponseWriter.Writer.Header().Set(hdr, val)
//	} else {
//		ctx.ResponseWriter.Writer.Header().Add(hdr, val)
//	}
//}

//func (ctx *Context) SetCookie(name string, value string, age int64) {
//	var utctime time.Time
//	if age == 0 {
//		// 2^31 - 1 seconds (roughly 2038)
//		utctime = time.Unix(2147483647, 0)
//	} else {
//		utctime = time.Unix(time.Now().Unix()+age, 0)
//	}
//	cookie := fmt.Sprintf("%s=%s; expires=%s", name, value, webTime(utctime))
//	ctx.ResponseWriter.AddHeader("Set-Cookie", cookie)
//}

//func webTime(t time.Time) string {
//	ftime := t.Format(time.RFC1123)
//	if strings.HasSuffix(ftime, "UTC") {
//		ftime = ftime[0:len(ftime)-3] + "GMT"
//	}
//	return ftime
//}
