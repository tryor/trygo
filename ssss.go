package ssss

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"time"
)

const VERSION = "0.0.1"

var (
	SSSSApp *App
	AppPath string
)

func init() {
	SSSSApp = NewApp(nil)
	AppPath, _ = os.Getwd()
}

type App struct {
	Handlers         *ControllerRegistor
	Config           *Config
	StaticDirs       map[string]string
	TemplateRegistor *TemplateRegistor
}

func NewApp(config *Config) *App {
	cr := NewControllerRegistor()
	app := &App{Handlers: cr,
		Config:           config,
		StaticDirs:       make(map[string]string),
		TemplateRegistor: NewTemplateRegistor()}
	cr.app = app
	return app
}

//method-http method, GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,*
//path-URL path
//name - method on the container
func (app *App) Register(method string, path string, c IController, name string, params ...string) *App {
	ps := ""
	if len(params) > 0 {
		ps = params[0]
	}
	app.Handlers.Add(method, path, c, name, ps)
	return app
}

//采用正则路由
func (app *App) RegisterPattern(method string, pattern string, c IController, name string, params ...string) *App {
	ps := ""
	if len(params) > 0 {
		ps = params[0]
	}
	app.Handlers.Add(method, pattern, c, name, ps, true)
	return app
}

func Register(method string, path string, c IController, name string, params ...string) *App {
	SSSSApp.Register(method, path, c, name, params...)
	return SSSSApp
}

func RegisterPattern(method string, path string, c IController, name string, params ...string) *App {
	SSSSApp.RegisterPattern(method, path, c, name, params...)
	return SSSSApp
}

func (app *App) SetStaticPath(url string, path string) *App {
	app.StaticDirs[url] = path
	return app
}

func SetStaticPath(url string, path string) *App {
	SSSSApp.StaticDirs[url] = path
	return SSSSApp
}

func AddTemplateExt(ext string) {
	SSSSApp.TemplateRegistor.AddTemplateExt(ext)
}

func AddFuncMap(key string, funname interface{}) error {
	return SSSSApp.TemplateRegistor.AddFuncMap(key, funname)
}

func (app *App) buildTemplate() {
	if app.Config.TemplatePath != "" {
		app.TemplateRegistor.buildTemplate(app.Config.TemplatePath)
	}
}

func (app *App) Run() {
	app.buildTemplate()
	if app.Config.HttpAddr == "" {
		app.Config.HttpAddr = "0.0.0.0"
	}

	//if app.config.FormatParamName == "" {
	//	app.config.FormatParamName = "_fmt"
	//}

	addr := fmt.Sprintf("%s:%d", app.Config.HttpAddr, app.Config.HttpPort)
	var err error

	for {
		if app.Config.UseFcgi {
			l, e := net.Listen("tcp", addr)
			if e != nil {
				log.Print("Listen: ", e)
			}
			//log.Print("UseFcgi, fcgi.Serve")
			err = fcgi.Serve(l, app.Handlers)
		} else {
			//log.Print("http.ListenAndServe")
			//err = http.ListenAndServe(addr, app.Handlers)
			err = httpListenAndServe(addr, app.Handlers, app.Config.ReadTimeout, app.Config.WriteTimeout)
		}
		if err != nil {
			log.Print("ListenAndServe: ", err)
			//panic(err)
		}
		time.Sleep(time.Second * 2)
	}
}

func httpListenAndServe(addr string, handler http.Handler, readTimeout time.Duration, writeTimeout time.Duration) error {
	//if readTimeout == 0 {
	//	readTimeout = time.Second * 5
	//}
	server := &http.Server{Addr: addr, Handler: handler, ReadTimeout: readTimeout, WriteTimeout: writeTimeout}
	return server.ListenAndServe()
}

func Run(config *Config) {
	SSSSApp.Config = config
	SSSSApp.Run()
}
