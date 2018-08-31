package trygo

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
	methods       []int8 //HTTP请求方法
	router        iRouter
	patternLength int
}

type funcParam struct {
	name     string
	isptr    bool //是否指针类型
	isstruct bool //是否结构类型
	typ      reflect.Type
}

func (p *funcParam) String() string {
	return p.name
}

func (p *funcParam) New() reflect.Value {
	if p.isptr {
		return reflect.New(p.typ)
	} else {
		return reflect.Indirect(reflect.New(p.typ))
	}
}

type iRouter interface {
	//是否自动解析请求参数
	ParseRequest(b bool) iRouter
	run(ctx *Context)
	parseRequestFlag() int8
}

const (
	parseRequestDefault = iota
	parseRequestYes
	parseRequestNo
)

type router struct {
	//0=def, 1=parse, 2=no parse
	parseRequest int8
}

func (r *router) run(ctx *Context) {
}

func (r *router) ParseRequest(b bool) iRouter {
	if b {
		r.parseRequest = parseRequestYes
	} else {
		r.parseRequest = parseRequestNo
	}
	return r
}

func (r *router) parseRequestFlag() int8 {
	return r.parseRequest
}

type defaultRouter struct {
	router
	app            *App
	controllerType reflect.Type //控制器类型
	funcName       string       //函数名称
	funcType       reflect.Type //函数类型
	//funcParamNames []string     //函数参数名称列表
	funcParams    []*funcParam //函数参数信息
	funcParamTags Taginfos     //参数的Tag信息
}

type restfulRouter struct {
	router
	handlerFunc HandlerFunc
}

type handlerRouter struct {
	router
	handler http.Handler
}

func (r *restfulRouter) run(ctx *Context) {
	r.handlerFunc(ctx)
}

func (r *handlerRouter) run(ctx *Context) {
	r.handler.ServeHTTP(ctx.ResponseWriter, ctx.Request)
}

func (r *defaultRouter) run(ctx *Context) {
	vc := reflect.New(r.controllerType)
	controller, ok := vc.Interface().(ControllerInterface)
	if !ok {
		panic(r.controllerType.String() + " is not ControllerInterface interface")
	}
	controller.Init(r.app, ctx, r.controllerType.Name(), r.funcName)

	defer controller.Finish()
	controller.Prepare()

	if r.funcName == "" {
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
		method := vc.MethodByName(r.funcName)
		numIn := r.funcType.NumIn()
		inx := make([]reflect.Value, numIn-1)
		if numIn > 1 {
			//auto bind func parameters
			tags := r.funcParamTags
			for i := 1; i < numIn; i++ {
				idx := i - 1
				funcParam := r.funcParams[idx]
				name := funcParam.name
				typ := funcParam.typ //router.funcType.In(i)
				v, err := ctx.Input.bind(name, typ, tags)
				if err != nil {
					ctx.Error = err
					if r.app.Config.ThrowBindParamPanic {
						var msg string
						if typ.Kind() == reflect.Struct {
							msg = fmt.Sprintf("%v, cause:%s.%v", ERROR_INFO_MAP[ERROR_CODE_PARAM_ILLEGAL], name, err)
						} else {
							msg = fmt.Sprintf("%v, %s=%v, cause:%v", ERROR_INFO_MAP[ERROR_CODE_PARAM_ILLEGAL], name, ctx.Input.Values[name], err)
						}
						panic(NewErrorResult(ERROR_CODE_PARAM_ILLEGAL, msg))
					}
					inx[idx] = funcParam.New()
				} else {
					if funcParam.isptr {
						inx[idx] = v.Addr()
					} else {
						inx[idx] = *v
					}
				}
			}
		}
		method.Call(inx)
	}

}

type ControllerRegister struct {
	explicitmap map[string]*controllerInfo
	relativemap map[string]*controllerInfo
	patternmap  map[*regexp.Regexp]*controllerInfo
	app         *App
	pool        sync.Pool
}

func NewControllerRegister(app *App) *ControllerRegister {
	cr := &ControllerRegister{
		app:         app,
		explicitmap: make(map[string]*controllerInfo),
		relativemap: make(map[string]*controllerInfo),
		patternmap:  make(map[*regexp.Regexp]*controllerInfo),
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
func (this *ControllerRegister) Add(methods string, pattern string, c ControllerInterface, name string, params []string, tags []string) (r iRouter) {
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

	routerinfo := &defaultRouter{app: this.app, funcName: name, controllerType: reflect.Indirect(controller).Type(), funcType: methodType, funcParams: make([]*funcParam, 0)}

	if params != nil && len(params) > 0 {
		for _, p := range params {
			paramname := strings.SplitN(strings.TrimSpace(p), " ", 2)[0]
			routerinfo.funcParams = append(routerinfo.funcParams, &funcParam{paramname, false, false, nil})
		}
	}

	methodParamTypes := make(map[string]reflect.Type) //key为参数名，值为参数类型//make([]reflect.Type, numIn, numIn)
	if methodType != nil {
		//check paramter num
		if methodType.NumIn()-1 != len(routerinfo.funcParams) {
			panic(fmt.Sprintf("http: the number of parameter mismatch, %v(%v), %v(%v)", routerinfo.funcParams, len(routerinfo.funcParams), methodType.String(), methodType.NumIn()-1))
		}

		numIn := methodType.NumIn()
		for i := 1; i < numIn; i++ {
			ptype := methodType.In(i)
			funcparam := routerinfo.funcParams[i-1]
			if ptype.Kind() == reflect.Ptr {
				ptype = ptype.Elem()
				funcparam.isptr = true
			}
			methodParamTypes[funcparam.name] = ptype
			//check struct
			if ptype.Kind() == reflect.Struct {
				for i := 0; i < ptype.NumField(); i++ {
					f := ptype.Field(i)
					firstChar := f.Name[0:1]
					if strings.ToLower(firstChar) == firstChar {
						panic(fmt.Sprintf("http: first char is lower, Struct:%v.%v", ptype.Name(), f.Name))
					}
				}
				funcparam.isstruct = true
			}
			funcparam.typ = ptype
		}

	}

	//parse tags
	routerinfo.funcParamTags = parseTags(methodParamTypes, tags, this.app.Config.FormDomainModel)
	if this.app.Config.RunMode == DEV {
		this.app.Logger.Debug("pattern:%v, action:%v.%v, tags:%v  => formatedTags:%v", pattern, reflect.TypeOf(c), name, tags, routerinfo.funcParamTags)
	}

	this.add(methods, pattern, routerinfo)
	return routerinfo
}

func (this *ControllerRegister) AddMethod(methods, pattern string, f HandlerFunc) (r iRouter) {
	r = &restfulRouter{handlerFunc: f}
	this.add(methods, pattern, r)
	return
}

func (this *ControllerRegister) AddHandler(pattern string, h http.Handler) (r iRouter) {
	r = &handlerRouter{handler: h}
	this.add("*", pattern, r)
	return
}

func (this *ControllerRegister) add(methods, pattern string, router iRouter) {

	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}

	if _, ok := this.explicitmap[pattern]; ok {
		panic("http: multiple registrations for " + pattern)
	}

	if _, ok := this.relativemap[pattern]; ok {
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
		controllerInfo.patternLength = len(pattern)
		this.patternmap[r] = controllerInfo
	} else {
		controllerInfo.patternLength = len(pattern)
		if pattern[len(pattern)-1] == '/' {
			this.relativemap[pattern] = controllerInfo
		} else {
			this.explicitmap[pattern] = controllerInfo
		}

	}
}

// usage:
//    Get("/", func(ctx *ssss.Context){
//          ...
//    })
func (this *ControllerRegister) Get(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("GET", pattern, f)
}

func (this *ControllerRegister) Post(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("POST", pattern, f)
}

func (this *ControllerRegister) Put(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("PUT", pattern, f)
}

func (this *ControllerRegister) Delete(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("DELETE", pattern, f)
}

func (this *ControllerRegister) Head(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("HEAD", pattern, f)
}

func (this *ControllerRegister) Patch(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("PATCH", pattern, f)
}

func (this *ControllerRegister) Options(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("OPTIONS", pattern, f)
}

func (this *ControllerRegister) Any(pattern string, f HandlerFunc) iRouter {
	return this.AddMethod("*", pattern, f)
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

	//if this.app.Config.RecoverFunc != nil {
	defer this.app.Config.RecoverFunc(ctx)
	//} else {
	//	defer defaultRecoverFunc(ctx)
	//}

	path := r.URL.Path
	pathlen := len(path)

	//http service router
	router, ok := this.explicitmap[path]
	if !ok {
		var n = 0
		for pattern, r := range this.relativemap {
			if !(pathlen >= r.patternLength && path[0:r.patternLength] == pattern) {
				continue
			}
			if router == nil || r.patternLength > n {
				n = r.patternLength
				router = r
			}
		}
	}
	if router != nil {
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
		Exec(true)
	if err != nil {
		this.app.Logger.Critical("%s", buildLoginfo(ctx.Request, err))
		//http.Error(ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
	}
}

func (this *ControllerRegister) call(routerinfo *controllerInfo, ctx *Context, restform url.Values) {

	preqflag := routerinfo.router.parseRequestFlag()
	if preqflag == parseRequestYes || (preqflag == parseRequestDefault && this.app.Config.AutoParseRequest) {
		err := ctx.Input.Parse()
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF || strings.HasSuffix(err.Error(), io.EOF.Error()) {
				this.app.Logger.Warn("client interrupt request, cause:%v", err)
				return
			} else if strings.Contains(err.Error(), ErrBodyTooLarge.Error()) {
				panic(NewErrorResult(ERROR_CODE_RUNTIME, err))
			} else {
				panic(err)
			}
		}
	}

	if restform != nil && len(restform) > 0 {
		ctx.Input.AddValues(restform)
	}

	routerinfo.router.run(ctx)
	render := ctx.ResponseWriter.render
	if render.started {
		err := render.Exec()
		if err != nil {
			ctx.App.Logger.Critical("%s", buildLoginfo(ctx.Request, err))
		}
	}
}

func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		return false
	}
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
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
			ctx.App.Logger.Info("%s", buildLoginfo(ctx.Request, errdata))
		case Result:
			errtxt = e.String()
			code = http.StatusBadRequest
			errdata = err
			//默认认为被Result包装过的异常为业务上的错误，采用Info日志级别
			ctx.App.Logger.Info("%s", buildLoginfo(ctx.Request, errdata))
		default:
			errtxt = "Internal Server Error"
			code = http.StatusInternalServerError
			errdata = fmt.Sprintf("%s, %v", errtxt, err)
			ctx.App.Logger.Critical("%s", buildLoginfo(ctx.Request, errdata))
			if ctx.App.Config.PrintPanicDetail || ctx.App.Config.RunMode == DEV {
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					ctx.App.Logger.Error("%s, %d", file, line)
				}
			}
		}

		err := Render(ctx, errdata).Status(code).
			Exec(true)

		if err != nil {
			ctx.App.Logger.Critical("%s", buildLoginfo(ctx.Request, err))
			//Error(ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		}

	}
}
