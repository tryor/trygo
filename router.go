package ssss

import (
	"sync"
	//	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const (
	ANY = iota
	GET
	POST
	PUT
	DELETE
	PATCH
	OPTIONS
	HEAD
	TRACE
	CONNECT
)

var (
	HttpMethods = map[string]int8{
		"*":       ANY,
		"GET":     GET,
		"POST":    POST,
		"PUT":     PUT,
		"DELETE":  DELETE,
		"PATCH":   PATCH,
		"OPTIONS": OPTIONS,
		"HEAD":    HEAD,
		"TRACE":   TRACE,
		"CONNECT": CONNECT,
	}
)

type HandlerFunc func(*Context)

type controllerInfo struct {
	methods []int8 //HTTP请求方法
	router  iRouter
}

type iRouter interface {
	Run(ctx *Context)
}

type defaultRouter struct {
	app            *App
	controllerType reflect.Type //控制器类型
	funcName       string       //函数名称
	funcType       reflect.Type //函数类型
	funcParamNames []string     //函数参数名称列表
	funcParamTags  Taginfos     //参数的Tag信息
}

type restfulRouter struct {
	handlerFunc HandlerFunc
}

type handlerRouter struct {
	handler http.Handler
}

func (router *restfulRouter) Run(ctx *Context) {
	router.handlerFunc(ctx)
}

func (router *handlerRouter) Run(ctx *Context) {
	router.handler.ServeHTTP(ctx.ResponseWriter, ctx.Request)
}

func (router *defaultRouter) Run(ctx *Context) {
	vc := reflect.New(router.controllerType)
	controller, ok := vc.Interface().(IController)
	if !ok {
		panic(router.controllerType.String() + " is not IController interface")
	}
	controller.Init(router.app, ctx, router.controllerType.Name(), router.funcName)

	controller.Prepare()
	defer controller.Finish()

	if router.funcName == "" {
		switch convMethod(ctx.Request.Method) {
		case GET:
			controller.Get()
		case POST:
			controller.Post()
		case PUT:
			controller.Put()
		case DELETE:
			controller.Delete()
		case PATCH:
			controller.Patch()
		case OPTIONS:
			controller.Options()
		case HEAD:
			controller.Head()
			//		case TRACE:
			//		case CONNECT:
		default:
			http.Error(ctx.ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
		}

	} else {
		method := vc.MethodByName(router.funcName)
		numIn := router.funcType.NumIn()
		inx := make([]reflect.Value, numIn-1)
		if numIn > 1 {
			//auto bind func parameters
			tags := router.funcParamTags
			for i := 1; i < numIn; i++ {
				idx := i - 1
				name := router.funcParamNames[idx]
				typ := router.funcType.In(i)
				//Logger.Debug("tags:%v", tags)
				v, err := ctx.Input.bind(name, typ, tags)
				if err != nil {
					ctx.Error = err
					if router.app.Config.ThrowBindParamPanic {
						var msg string
						if typ.Kind() == reflect.Struct {
							msg = fmt.Sprintf("%v, cause:%s.%v", ERROR_INFO_MAP[ERROR_CODE_PARAM_ILLEGAL], name, err)
						} else {
							msg = fmt.Sprintf("%v, %s=%v, cause:%v", ERROR_INFO_MAP[ERROR_CODE_PARAM_ILLEGAL], name, ctx.Input.Values[name], err)
						}
						panic(NewErrorResult(ERROR_CODE_PARAM_ILLEGAL, msg))
						//panic(NewErrorResult(ERROR_CODE_PARAM_ILLEGAL, fmt.Sprintf("%v=%v,cause:%v", name, ctx.Input.Values[name], err)))
					}
					inx[idx] = reflect.Indirect(reflect.New(typ))
				} else {
					inx[idx] = *v
				}
			}
		}
		method.Call(inx)
	}

}

type ControllerRegister struct {
	routermap  map[string]*controllerInfo
	patternmap map[*regexp.Regexp]*controllerInfo
	app        *App
	pool       sync.Pool
}

func NewControllerRegister() *ControllerRegister {
	cr := &ControllerRegister{
		routermap:  make(map[string]*controllerInfo),
		patternmap: make(map[*regexp.Regexp]*controllerInfo),
	}
	cr.pool.New = func() interface{} {
		return newContext()
	}
	return cr
}

//method - http method, GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,*
//path- URL path
//name - method on the container
//params - parameter name list
//tags parameter tag info
func (this *ControllerRegister) Add(methods string, pattern string, c IController, name string, params []string, tags []string) {
	if c == nil {
		panic("http: controller is empty")
	}

	var methodType reflect.Type
	controller := reflect.ValueOf(c)
	if name != "" {
		controllerMethod := controller.MethodByName(name)
		controllerType := reflect.TypeOf(c)
		controllerTypeMethod, ok := controllerType.MethodByName(name)
		if !controllerMethod.IsValid() && !ok {
			panic(fmt.Sprintf("http: ROUTER METHOD [%v] not find or invalid", name))
		}
		methodType = controllerTypeMethod.Type
	}

	routerinfo := &defaultRouter{app: this.app, funcName: name, controllerType: reflect.Indirect(controller).Type(), funcType: methodType, funcParamNames: make([]string, 0)}

	if params != nil && len(params) > 0 {
		for _, p := range params {
			routerinfo.funcParamNames = append(routerinfo.funcParamNames, strings.SplitN(strings.TrimSpace(p), " ", 2)[0])
		}
	}

	methodParamTypes := make(map[string]reflect.Type) //key为参数名，值为参数类型//make([]reflect.Type, numIn, numIn)
	if methodType != nil {
		//check paramter num
		if methodType.NumIn()-1 != len(routerinfo.funcParamNames) {
			panic(fmt.Sprintf("http: the number of parameter mismatch, %v(%v), %v(%v)", routerinfo.funcParamNames, len(routerinfo.funcParamNames), methodType.String(), methodType.NumIn()-1))
		}

		numIn := methodType.NumIn()
		for i := 1; i < numIn; i++ {
			ptype := methodType.In(i)
			methodParamTypes[routerinfo.funcParamNames[i-1]] = ptype
			//check struct
			if ptype.Kind() == reflect.Struct {
				for i := 0; i < ptype.NumField(); i++ {
					f := ptype.Field(i)
					firstChar := f.Name[0:1]
					if strings.ToLower(firstChar) == firstChar {
						panic(fmt.Sprintf("http: first char is lower, Struct:%v.%v", ptype.Name(), f.Name))
					}
				}
			}
		}

	}

	//parse tags
	routerinfo.funcParamTags = parseTags(methodParamTypes, tags, this.app.Config.FormDomainModel)
	if this.app.Config.RunMode == DEV {
		Logger.Debug("pattern:%v, action:%v.%v, tags:%v  => formatedTags:%v", pattern, reflect.TypeOf(c), name, tags, routerinfo.funcParamTags)
	}

	this.add(methods, pattern, routerinfo)
}

func (this *ControllerRegister) AddMethod(methods, pattern string, f HandlerFunc) {
	this.add(methods, pattern, &restfulRouter{f})
}

func (this *ControllerRegister) AddHandler(pattern string, h http.Handler) {
	this.add("*", pattern, &handlerRouter{h})
}

func (this *ControllerRegister) add(methods, pattern string, router iRouter) {

	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}

	if _, ok := this.routermap[pattern]; ok {
		panic("http: multiple registrations for " + pattern)
	}

	for r, _ := range this.patternmap {
		if r.String() == pattern {
			panic("http: multiple registrations for " + pattern)
		}
	}

	httpMethodstrs := strings.Split(methods, "|")
	httpMethods := make([]int8, 0, len(httpMethodstrs))
	for _, m := range httpMethodstrs {
		mv := convMethod(m)
		if mv == 0 {
			for _, v := range HttpMethods {
				httpMethods = append(httpMethods, v)
			}
		} else {
			httpMethods = append(httpMethods, mv)
		}
	}

	if len(httpMethods) == 0 {
		panic("http: methods is empty")
	}

	controllerInfo := &controllerInfo{methods: httpMethods, router: router}

	if isPattern(pattern) {
		//regex router
		if pattern[0] != '^' && pattern[0] == '/' {
			pattern = "^" + pattern
		}
		r, err := regexp.Compile(pattern)
		if err != nil {
			panic(err)
		}
		this.patternmap[r] = controllerInfo
	} else {
		this.routermap[pattern] = controllerInfo
	}
}

// usage:
//    Get("/", func(ctx *ssss.Context){
//          ...
//    })
func (this *ControllerRegister) Get(pattern string, f HandlerFunc) {
	this.AddMethod("GET", pattern, f)
}

func (this *ControllerRegister) Post(pattern string, f HandlerFunc) {
	this.AddMethod("POST", pattern, f)
}

func (this *ControllerRegister) Put(pattern string, f HandlerFunc) {
	this.AddMethod("PUT", pattern, f)
}

func (this *ControllerRegister) Delete(pattern string, f HandlerFunc) {
	this.AddMethod("DELETE", pattern, f)
}

func (this *ControllerRegister) Head(pattern string, f HandlerFunc) {
	this.AddMethod("HEAD", pattern, f)
}

func (this *ControllerRegister) Patch(pattern string, f HandlerFunc) {
	this.AddMethod("PATCH", pattern, f)
}

func (this *ControllerRegister) Options(pattern string, f HandlerFunc) {
	this.AddMethod("OPTIONS", pattern, f)
}

func (this *ControllerRegister) Any(pattern string, f HandlerFunc) {
	this.AddMethod("*", pattern, f)
}

func (this *ControllerRegister) hasHttpMethod(router *controllerInfo, method int8) bool {
	for _, m := range router.methods {
		if method == m {
			return true
		}
	}
	return false
}

//*,GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,TRACE,CONNECT => 0,1,2,3,4,5,6,7,8,9
func convMethod(m string) int8 {
	mv, ok := HttpMethods[strings.ToUpper(m)]
	if ok {
		return mv
	}
	panic("(" + m + ") Method is not supported")
}

func (this *ControllerRegister) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := this.pool.Get().(*Context)
	ctx.Reset(rw, r, this.app)
	defer this.pool.Put(ctx)

	if this.app.Config.RecoverFunc != nil {
		defer this.app.Config.RecoverFunc(ctx)
	} else {
		defer defaultRecoverFunc(ctx)
	}

	path := r.URL.Path

	//http service route
	router, ok := this.routermap[path]
	if ok {
		if this.hasHttpMethod(router, convMethod(r.Method)) {
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

		if this.hasHttpMethod(router, convMethod(r.Method)) {
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

	err := Render(ctx, "Method Not Allowed").Text().
		Status(http.StatusMethodNotAllowed).
		Exec()
	if err != nil {
		Logger.Critical("%v", err)
		http.Error(ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
	}
}

func (this *ControllerRegister) call(routerinfo *controllerInfo, ctx *Context, restform url.Values) {
	if this.app.Config.AutoParseRequest {
		err := ctx.Input.Parse()
		if err != nil {
			panic(err)
		}
	}

	if restform != nil && len(restform) > 0 {
		ctx.Input.AddValues(restform)
	}

	routerinfo.router.Run(ctx)
	render := ctx.ResponseWriter.render
	if render.started {
		err := render.Exec()
		if err != nil {
			panic(err)
		}
	}
}

func defaultRecoverFunc(ctx *Context) {
	if err := recover(); err != nil {
		var errtxt string
		var code int
		var errdata interface{}

		switch e := err.(type) {
		case *Result:
			errtxt = e.String()
			code = http.StatusBadRequest
			errdata = err
			//默认认为被Result包装过的异常为业务上的错误，采用Info日志级别
			Logger.Info("%v", errtxt)
		case Result:
			errtxt = e.String()
			code = http.StatusBadRequest
			errdata = err
			//默认认为被Result包装过的异常为业务上的错误，采用Info日志级别
			Logger.Info("%v", errtxt)
		default:
			errtxt = "Internal Server Error"
			code = http.StatusInternalServerError
			errdata = fmt.Sprintf("%s, %v", errtxt, err)
			Logger.Critical("%v", errdata)

			if ctx.App.Config.PrintPanicDetail || ctx.App.Config.RunMode == DEV {
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					Logger.Error("%s, %d", file, line)
				}
			}
		}

		err := Render(ctx, errdata).Status(code).
			Exec()

		if err != nil {
			Logger.Critical("%v", err)
			http.Error(ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		}

	}
}
