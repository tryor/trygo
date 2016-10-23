package ssss

import (
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

type controllerInfo struct {
	methods        []int8 //HTTP方法
	all            bool
	controllerType reflect.Type
	name           string       //函数名称
	typ            reflect.Type //函数类型
	pnames         []string     //函数参数名称列表
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
func (this *ControllerRegistor) Add(methods string, path string, c IController, name string, params string, regex ...bool) {
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

	//检查参数类型
	for i := 1; i < mtype.Type.NumIn(); i++ {
		ptype := mtype.Type.In(i)
		if (ptype.Kind() != reflect.String && ptype.Kind() != reflect.Slice) || (ptype.Kind() == reflect.Slice && ptype.Elem().Kind() != reflect.String) {
			panic(fmt.Sprintf("the parameter type is not string, %v", ptype))
		}
	}

	ms := strings.Split(methods, "|")
	routerinfo := &controllerInfo{methods: make([]int8, len(ms)), all: false, name: name, controllerType: reflect.Indirect(reflect.ValueOf(c)).Type(), typ: mtype.Type, pnames: make([]string, 0)}
	params = strings.TrimSpace(params)
	if params != "" {
		for _, p := range strings.Split(params, ",") {
			routerinfo.pnames = append(routerinfo.pnames, strings.SplitN(strings.TrimSpace(p), " ", 2)[0])
		}
	}

	if mtype.Type.NumIn()-1 != len(routerinfo.pnames) {
		panic(fmt.Sprintf("the number of parameter mismatch, %v(%v), %v(%v)", routerinfo.pnames, len(routerinfo.pnames), mtype.Type.String(), mtype.Type.NumIn()-1))
	}

	for i, m := range ms {
		routerinfo.methods[i] = this.convMethod(strings.ToUpper(m))
		if routerinfo.methods[i] == 0 {
			routerinfo.all = true
		}
	}
	if len(routerinfo.methods) == 0 {
		panic("methods is empty")
	}
	//log.Debugf("ROUTER PATH [%v] METHOD [%v]", path, name)

	//fmt.Println(regex)
	if len(regex) == 0 || !regex[0] {
		this.routermap[path] = routerinfo
	} else {
		//regex router
		r, err := regexp.Compile(path)
		if err != nil {
			panic(err)
		}
		this.patternmap[r] = routerinfo
		//fmt.Println(r)
	}
}

// AutoRoute
func (this *ControllerRegistor) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Internal Server Error, %v", err)
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			if this.app.Config.PrintPanic {
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Print(file, line)
				}
			}
		}
	}()

	path := r.URL.Path

	//http service route
	router, ok := this.routermap[path]
	if ok && router != nil {
		if router.all || this.hasMethod(router, this.convMethod(r.Method)) {
			this.call(router, rw, r)
			return
		}
	}

	//regex router
	for regex, router := range this.patternmap {
		if !regex.MatchString(path) {
			continue
		}
		if router.all || this.hasMethod(router, this.convMethod(r.Method)) {
			this.call(router, rw, r)
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

func (this *ControllerRegistor) call(router *controllerInfo, rw http.ResponseWriter, r *http.Request) {
	vc := reflect.New(router.controllerType)

	init := vc.MethodByName("Init")
	in := make([]reflect.Value, 4)
	ct := &Context{ResponseWriter: rw, Request: r}
	in[0] = reflect.ValueOf(this.app)
	in[1] = reflect.ValueOf(ct)
	in[2] = reflect.ValueOf(router.controllerType.Name())
	in[3] = reflect.ValueOf(router.name)
	init.Call(in)

	in0 := make([]reflect.Value, 0)
	method := vc.MethodByName("Prepare")
	if !method.Call(in0)[0].Interface().(bool) {
		return
	}

	method = vc.MethodByName(router.name)
	numIn := router.typ.NumIn()
	inx := make([]reflect.Value, numIn-1)
	if numIn > 1 {
		parseForm := vc.MethodByName("ParseForm")
		res := parseForm.Call(in0)
		err := res[1].Interface()
		if err != nil {
			panic(err.(error))
		}
		form := res[0].Interface().(url.Values)
		for i := 1; i < numIn; i++ {
			idx := i - 1
			inx[idx] = reflect.ValueOf(this.parseParam(form, router.pnames[idx], router.typ.In(i)))
		}
	}

	defer func() {
		panicInfoField := vc.Elem().FieldByName("PanicInfo")
		er := recover()
		if er != nil {
			panicInfoField.Set(reflect.ValueOf(er))
		}
		method = vc.MethodByName("Finish")
		method.Call(in0)
		if er != nil && !panicInfoField.IsNil() {
			panic(er)
		}
	}()
	method.Call(inx)
}

func (this *ControllerRegistor) parseParam(form url.Values, pname string, ptype reflect.Type) interface{} {
	vp := reflect.Indirect(reflect.New(ptype))

	if reflect.Slice == ptype.Kind() {
		kind := ptype.Elem().Kind()
		vals := form[pname]
		for _, str := range vals {
			v := reflect.Indirect(reflect.New(ptype.Elem()))
			parseValue(kind, str, &v)
			vp = reflect.Append(vp, v)
		}
	} else {
		parseValue(ptype.Kind(), form.Get(pname), &vp)
	}

	return vp.Interface()
}

//在vp中返回值
func parseValue(kind reflect.Kind, val string, vp *reflect.Value) {

	switch kind {
	case reflect.String:
		vp.SetString(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(getValue(val, "false"))
		if err != nil {
			panic(err)
		}
		vp.SetBool(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(getValue(val, "0"), 10, 0)
		if err != nil {
			panic(err)
		}
		vp.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(getValue(val, "0"), 10, 0)
		if err != nil {
			panic(err)
		}
		vp.SetUint(val)
	case reflect.Float32:
		val, err := strconv.ParseFloat(getValue(val, "0.0"), 32)
		if err != nil {
			panic(err)
		}
		vp.SetFloat(val)
	case reflect.Float64:
		val, err := strconv.ParseFloat(getValue(val, "0.0"), 64)
		if err != nil {
			panic(err)
		}
		vp.SetFloat(val)
	default:
		panic("the parameter type is not supported")
	}
}

func getValue(val string, def string) string {
	if val == "" {
		return def
	}
	return val
}

func (this *ControllerRegistor) hasMethod(router *controllerInfo, method int8) bool {
	for _, m := range router.methods {
		if method == m {
			return true
		}
	}
	return false
}

//*,GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS => 0,1,2,3,4,5,6,7,8
func (this *ControllerRegistor) convMethod(m string) int8 {
	switch m {
	case "*":
		return 0
	case "GET":
		return 1
	case "POST":
		return 2
	case "PUT":
		return 3
	case "HEAD":
		return 4
	case "DELETE":
		return 5
	case "PATCH":
		return 6
	case "OPTIONS":
		return 7
	}
	panic("(" + m + ") Method is not supported")
}
