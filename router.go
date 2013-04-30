package ssss

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type controllerInfo struct {
	methods        []int8
	all            bool
	controllerType reflect.Type
	name           string
}

type ControllerRegistor struct {
	routermap map[string]*controllerInfo
	app       *App
}

func NewControllerRegistor() *ControllerRegistor {
	return &ControllerRegistor{routermap: make(map[string]*controllerInfo)}
}

//method-http method, GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,*
//path-URL path
//name - method on the container
func (this *ControllerRegistor) Add(methods string, path string, c IController, name string) {
	if c == nil {
		panic("controller is empty")
	}
	if name == "" {
		panic("method name on the container is empty")
	}

	appntv := reflect.ValueOf(c)
	m := appntv.MethodByName(name)
	if !m.IsValid() {
		panic(fmt.Sprintf("ROUTER METHOD [%v] not find or invalid", name))
	}

	ms := strings.Split(methods, "|")
	routerinfo := &controllerInfo{methods: make([]int8, len(ms)), all: false, name: name, controllerType: reflect.Indirect(reflect.ValueOf(c)).Type()}

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

	this.routermap[path] = routerinfo
}

// AutoRoute
func (this *ControllerRegistor) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			if !this.app.config.RecoverPanic {
				panic(err)
			} else {
				log.Print("Handler crashed with error,", err)
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Print(file, line)
				}
				http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			}
		}
	}()

	//http service route
	router, ok := this.routermap[r.URL.Path]
	if ok && router != nil {
		if router.all || this.hasMethod(router, this.convMethod(r.Method)) {
			vc := reflect.New(router.controllerType)

			init := vc.MethodByName("Init")
			in := make([]reflect.Value, 4)
			ct := &Context{ResponseWriter: rw, Request: r}
			in[0] = reflect.ValueOf(this.app)
			in[1] = reflect.ValueOf(ct)
			in[2] = reflect.ValueOf(router.controllerType.Name())
			in[3] = reflect.ValueOf(router.name)
			init.Call(in)

			in = make([]reflect.Value, 0)
			method := vc.MethodByName("Prepare")
			method.Call(in)

			method = vc.MethodByName(router.name)
			method.Call(in)

			method = vc.MethodByName("Finish")
			method.Call(in)
			return
		}
	}

	//static file server
	for p, dir := range this.app.StaticDirs {
		if strings.HasPrefix(r.URL.Path, p) {
			var file string
			if p == "/" {
				file = dir + r.URL.Path
			} else {
				file = dir + r.URL.Path[len(p):]
			}
			http.ServeFile(rw, r, file)
			return
		}
	}

	http.Error(rw, "Method Not Allowed", http.StatusMethodNotAllowed)
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
