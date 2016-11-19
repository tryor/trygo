package ssss

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	HttpMethods = map[string]int8{
		"*":       0,
		"GET":     1,
		"POST":    2,
		"PUT":     3,
		"DELETE":  4,
		"PATCH":   5,
		"OPTIONS": 6,
		"HEAD":    7,
		"TRACE":   8,
		"CONNECT": 9,
	}
)

type controllerInfo struct {
	methods []int8 //HTTP方法
	//any            bool
	controllerType reflect.Type
	name           string              //函数名称
	typ            reflect.Type        //函数类型
	pnames         []string            //函数参数名称列表
	tags           map[string]*tagInfo //参数的Tag信息
}

type ControllerRegistor struct {
	routermap  map[string]*controllerInfo
	patternmap map[*regexp.Regexp]*controllerInfo //key为编译后的正则模式
	app        *App
}

func NewControllerRegistor() *ControllerRegistor {
	return &ControllerRegistor{routermap: make(map[string]*controllerInfo), patternmap: make(map[*regexp.Regexp]*controllerInfo)}
}

//method - http method, GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,*
//path- URL path
//name - method on the container
//params - parameter name list
//tags parameter tag info
func (this *ControllerRegistor) Add(methods string, path string, c IController, name string, params []string, tags []string) {
	if c == nil {
		panic("controller is empty")
	}
	if name == "" {
		panic("method name on the container is empty")
	}

	appntv := reflect.ValueOf(c)
	m := appntv.MethodByName(name)
	ctype := reflect.TypeOf(c)
	mtype, ok := ctype.MethodByName(name)
	if !m.IsValid() && !ok {
		panic(fmt.Sprintf("ROUTER METHOD [%v] not find or invalid", name))
	}

	//	检查参数类型
	//	for i := 1; i < mtype.Type.NumIn(); i++ {
	//		ptype := mtype.Type.In(i)
	//		if (ptype.Kind() != reflect.String && ptype.Kind() != reflect.Slice) || (ptype.Kind() == reflect.Slice && ptype.Elem().Kind() != reflect.String) {
	//			panic(fmt.Sprintf("the parameter type is not string, %v", ptype))
	//		}
	//	}

	httpMethods := strings.Split(methods, "|")
	routerinfo := &controllerInfo{methods: make([]int8, 0, len(httpMethods)), name: name, controllerType: reflect.Indirect(reflect.ValueOf(c)).Type(), typ: mtype.Type, pnames: make([]string, 0)}
	//params = strings.TrimSpace(params)
	if params != nil && len(params) > 0 { //params != "" {
		for _, p := range params { //strings.Split(params, ",") {
			routerinfo.pnames = append(routerinfo.pnames, strings.SplitN(strings.TrimSpace(p), " ", 2)[0])
		}
	}

	//check paramter num
	if mtype.Type.NumIn()-1 != len(routerinfo.pnames) {
		panic(fmt.Sprintf("the number of parameter mismatch, %v(%v), %v(%v)", routerinfo.pnames, len(routerinfo.pnames), mtype.Type.String(), mtype.Type.NumIn()-1))
	}

	numIn := mtype.Type.NumIn()
	methodParamTypes := make(map[string]reflect.Type) //key为参数名，值为参数类型//make([]reflect.Type, numIn, numIn)
	for i := 1; i < numIn; i++ {
		ptype := mtype.Type.In(i)
		methodParamTypes[routerinfo.pnames[i-1]] = ptype
		//check struct
		if ptype.Kind() == reflect.Struct {
			for i := 0; i < ptype.NumField(); i++ {
				f := ptype.Field(i)
				firstChar := f.Name[0:1]
				if strings.ToLower(firstChar) == firstChar {
					panic(fmt.Sprintf("first char is lower, Struct:%v.%v", ptype.Name(), f.Name))
				}
			}
		}
	}

	for _, m := range httpMethods {
		mv := this.convMethod(m)
		if mv == 0 {
			//routerinfo.any = true
			for _, v := range HttpMethods {
				routerinfo.methods = append(routerinfo.methods, v)
			}
		} else {
			routerinfo.methods = append(routerinfo.methods, mv)
		}
	}

	if len(routerinfo.methods) == 0 {
		panic("methods is empty")
	}

	//parse tags
	routerinfo.tags = parseTags(methodParamTypes, tags, this.app.Config.FormDomainModel)

	if isPattern(path) {
		//regex router
		r, err := regexp.Compile(path)
		if err != nil {
			panic(err)
		}
		this.patternmap[r] = routerinfo
	} else {
		this.routermap[path] = routerinfo
	}
}

func defaultRecoverFunc(ctx *Context) {
	if err := recover(); err != nil {
		var errtxt string
		var code int
		var errdata interface{}

		switch err.(type) {
		case *Result, Result:
			if e, ok := err.(Result); ok {
				errtxt = e.String()
			} else {
				errtxt = fmt.Sprint(err)
			}
			code = http.StatusBadRequest
			errdata = err
			Logger.Error("%v", errtxt)
		default:
			errtxt = "Internal Server Error"
			code = http.StatusInternalServerError
			errdata = err
			Logger.Critical("%v, %v", errtxt, err)

			if ctx.config.PrintPanic {
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Print(file, line)
				}
			}
		}

		if ctx.config.ResponseFormatPanic {
			format, jsoncallback := ctx.Request.FormValue("fmt"), ctx.Request.FormValue("jsoncallback")
			data, err := BuildError(errdata, format, jsoncallback)
			if err != nil {
				Logger.Error("%v", err)
				http.Error(ctx.ResponseWriter, errtxt, code)
			} else {
				RenderData(ctx.ResponseWriter, format, data)
			}
		} else {
			http.Error(ctx.ResponseWriter, errtxt, code)
		}
	}
}

func (this *ControllerRegistor) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := &Context{ResponseWriter: rw, Request: r, config: this.app.Config}

	if this.app.Config.RecoverFunc != nil {
		defer this.app.Config.RecoverFunc(ctx)
	} else {
		defer defaultRecoverFunc(ctx)
	}

	path := r.URL.Path
	ctx.Multipart = strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data")

	//http service route
	router, ok := this.routermap[path]
	if ok { // && router != nil {
		//if router.any || this.hasMethod(router, this.convMethod(r.Method)) {
		if this.hasMethod(router, this.convMethod(r.Method)) {
			this.call(router, ctx, nil)
			return
		}
	}

	//regex router
	for regex, router := range this.patternmap {
		items := regex.FindStringSubmatch(path)
		if items == nil || len(items) == 0 {
			continue
		}
		restform := make(url.Values)
		for i, name := range regex.SubexpNames() {
			if i == 0 {
				continue
			}
			if name == "" {
				restform.Add("$"+strconv.Itoa(i), items[i])
			} else {
				restform.Add(name, items[i])
			}
		}

		//if router.any || this.hasMethod(router, this.convMethod(r.Method)) {
		if this.hasMethod(router, this.convMethod(r.Method)) {
			this.call(router, ctx, restform)
			return
		}
	}

	//static file server
	for p, dir := range this.app.StaticDirs {
		if strings.HasPrefix(r.URL.Path, p) {
			var file string
			if p == "/" {
				file = dir + path
			} else {
				file = dir + path[len(p):]
			}
			http.ServeFile(rw, r, file)
			return
		}
	}

	http.Error(rw, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (this *ControllerRegistor) call(router *controllerInfo, ctx *Context, restform url.Values) {
	vc := reflect.New(router.controllerType)

	controller, ok := vc.Interface().(IController)
	if !ok {
		panic(router.controllerType.String() + " is not IController interface")
	}

	controller.Init(this.app, ctx, router.controllerType.Name(), router.name)

	form, err := controller.ParseForm()
	if err != nil {
		panic(err)
	}

	if restform != nil {
		for k, v := range restform {
			form[k] = append(form[k], v...)
		}
	}

	controller.Prepare()
	defer controller.Finish()

	method := vc.MethodByName(router.name)
	numIn := router.typ.NumIn()
	inx := make([]reflect.Value, numIn-1)
	if numIn > 1 {
		tags := router.tags
		for i := 1; i < numIn; i++ {
			idx := i - 1
			pname := router.pnames[idx]
			v, err := this.parseMethodParam(form, tags, pname, router.typ.In(i))
			if err != nil {
				panic(NewErrorResult(ERROR_CODE_PARAM_ILLEGAL, fmt.Sprintf("%v=%v,cause:%v", pname, form[pname], err)))
			} else {
				inx[idx] = *v
			}
		}
	}
	method.Call(inx)
}

func (this *ControllerRegistor) parseMethodParam(form url.Values, tags map[string]*tagInfo, pname string, ptype reflect.Type) (*reflect.Value, error) {
	vp := reflect.Indirect(reflect.New(ptype))
	kind := ptype.Kind()
	switch kind {
	case reflect.Slice:
		kind = ptype.Elem().Kind()
		if reflect.Struct == kind {
			return nil, errors.New("the parameter slice type is not supported")
		} else {
			vals := form[pname]
			for _, str := range vals {
				v := reflect.Indirect(reflect.New(ptype.Elem()))
				if err := this.parseValue(tags[pname], kind, str, true, &v); err != nil {
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
				paramTag := f.Tag.Get("field")
				paramTags := strings.SplitN(paramTag, ",", 2)
				if len(paramTag) > 0 {
					name = strings.TrimSpace(paramTags[0])
				}
				if len(name) == 0 {
					name = strings.ToLower(f.Name[0:1]) + f.Name[1:]
				}
				//if _, ok := form[name]; !ok {
				//	name = f.Name
				//}
				if this.app.Config.FormDomainModel {
					name = pname + "." + name
				}
			}
			v, err := this.parseMethodParam(form, tags, name, f.Type)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%v=%v, %v", name, form[name], err))
			}
			vp.Field(i).Set(*v)
		}

	case reflect.Map:
		vp.Set(reflect.ValueOf(form))

	default:
		vals, ok := form[pname]
		var val string
		if ok {
			val = vals[0]
		}
		if err := this.parseValue(tags[pname], ptype.Kind(), val, ok, &vp); err != nil {
			return nil, err
		}
	}
	return &vp, nil
}

//在vp中返回值
func (this *ControllerRegistor) parseValue(tagInfo *tagInfo, kind reflect.Kind, val string, ok bool, vp *reflect.Value) error {
	if tagInfo != nil {
		//default value
		if tagInfo.Default.Exist {
			if kind == reflect.String {
				if !ok {
					vp.Set(tagInfo.Default.Value)
					return nil
				}
			} else if len(val) == 0 {
				vp.Set(tagInfo.Default.Value)
				return nil
			}
		}

		//check require
		if tagInfo.Require && (!ok || len(val) == 0) {
			return errors.New("value is empty, require")
		}

		//check limit
		if tagInfo.Limit.Exist && tagInfo.Limit.Value > 0 {
			if len(val) > tagInfo.Limit.Value {
				return errors.New(fmt.Sprintf("value is too long, limit:%v", tagInfo.Limit.Value))
			}
		}

		//check pattern
		if tagInfo.Pattern.Exist && len(val) > 0 {
			if !tagInfo.Pattern.Regexp.MatchString(val) {
				return errors.New(fmt.Sprintf("value is illegal, pattern match fail!, pattern:%v", tagInfo.Pattern.Regexp.String()))
			}
		}
	}

	switch kind {
	case reflect.String:
		vp.SetString(val)
	case reflect.Bool:
		if len(val) > 0 {
			bval, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			vp.SetBool(bval)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if len(val) > 0 {
			ival, err := strconv.ParseInt(val, 10, 0)
			if err != nil {
				return err
			}
			vp.SetInt(ival)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if len(val) > 0 {
			uval, err := strconv.ParseUint(val, 10, 0)
			if err != nil {
				return err
			}
			vp.SetUint(uval)
		}
	case reflect.Float32:
		if len(val) > 0 {
			fval, err := strconv.ParseFloat(val, 32)
			if err != nil {
				return err
			}
			vp.SetFloat(fval)
		}
	case reflect.Float64:
		if len(val) > 0 {
			fval, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return err
			}
			vp.SetFloat(fval)
		}
	default:
		return errors.New(fmt.Sprintf("the parameter type is not supported, type is %v, value is %v", vp.Type(), val))
	}

	if len(val) > 0 && tagInfo != nil && tagInfo.Scope.Exist {
		if !tagInfo.Scope.Items.check(vp.Interface()) {
			return errors.New(fmt.Sprintf("value is illegal, scope:[%v]", tagInfo.Scope.String()))
		}
	}

	return nil
}

func (this *ControllerRegistor) hasMethod(router *controllerInfo, method int8) bool {
	for _, m := range router.methods {
		if method == m {
			return true
		}
	}
	return false
}

//*,GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,TRACE,CONNECT => 0,1,2,3,4,5,6,7,8,9
func (this *ControllerRegistor) convMethod(m string) int8 {
	mv, ok := HttpMethods[strings.ToUpper(m)]
	if ok {
		return mv
	}
	panic("(" + m + ") Method is not supported")
}
