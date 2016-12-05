package ssss

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"reflect"
	"strings"
	//	"time"
)

type App struct {
	Handlers         *ControllerRegister
	Config           *config
	StaticDirs       map[string]string
	TemplateRegister *TemplateRegister
}

func NewApp() *App {
	cr := NewControllerRegister()
	app := &App{Handlers: cr,
		Config:           newConfig(),
		StaticDirs:       make(map[string]string),
		TemplateRegister: NewTemplateRegister()}
	cr.app = app
	return app
}

//method - http method, GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,*
//pattern - URL path or regexp pattern
//name - method on the container
//tags - function parameter tag info, see struct tag
func (app *App) Register(method string, pattern string, c IController, name string, tags ...string) *App {
	funcname, params := parseMethod(name)
	app.Handlers.Add(method, pattern, c, funcname, params, tags)
	return app
}

func (app *App) RegisterHandler(pattern string, h http.Handler) *App {
	app.Handlers.AddHandler(pattern, h)
	return app
}

func (app *App) RegisterFunc(methods, pattern string, f HandlerFunc) *App {
	app.Handlers.AddMethod(methods, pattern, f)
	return app
}

func (app *App) RegisterRESTful(pattern string, c IController) *App {
	app.Register("*", pattern, c, "")
	if !isPattern(pattern) {
		//pattern = path.Join("^", pattern, "(?P<id>[^/]+)/?$")
		pattern = path.Join(pattern, "(?P<id>[^/]+)$")
	}
	app.Register("*", pattern, c, "")
	return app
}

func (app *App) Get(pattern string, f HandlerFunc) *App {
	app.Handlers.Get(pattern, f)
	return app
}

func (app *App) Post(pattern string, f HandlerFunc) *App {
	app.Handlers.Post(pattern, f)
	return app
}

func (app *App) Put(pattern string, f HandlerFunc) *App {
	app.Handlers.Put(pattern, f)
	return app
}

func (app *App) Delete(pattern string, f HandlerFunc) *App {
	app.Handlers.Delete(pattern, f)
	return app
}

func (app *App) Head(pattern string, f HandlerFunc) *App {
	app.Handlers.Head(pattern, f)
	return app
}

func (app *App) Patch(pattern string, f HandlerFunc) *App {
	app.Handlers.Patch(pattern, f)
	return app
}

func (app *App) Options(pattern string, f HandlerFunc) *App {
	app.Handlers.Options(pattern, f)
	return app
}

func (app *App) Any(pattern string, f HandlerFunc) *App {
	app.Handlers.Any(pattern, f)
	return app
}

func (app *App) SetStaticPath(url string, path string) *App {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if url != "/" {
		url = strings.TrimRight(url, "/")
	}
	app.StaticDirs[url] = path
	return app
}

func (app *App) SetViewsPath(path string) *App {
	app.Config.TemplatePath = path
	return app
}

func (app *App) buildTemplate() {
	if app.Config.TemplatePath != "" {
		err := app.TemplateRegister.buildTemplate(app.Config.TemplatePath)
		if err != nil {
			Logger.Error("%v", err)
		}
	}
}

func (app *App) Run(listener ...HttpServeListener) {
	app.buildTemplate()
	if app.Config.HttpAddr == "" {
		app.Config.HttpAddr = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%d", app.Config.HttpAddr, app.Config.HttpPort)
	var err error
	var hsl HttpServeListener
	if len(listener) > 0 {
		hsl = listener[0]
	} else {
		hsl = &DefaultHttpServeListener{}
	}

	if err = hsl.ListenAndServe(app, addr, app.Handlers); err != nil {
		Logger.Critical("%v.ListenAndServe: %v", reflect.TypeOf(hsl), err)
	}

}

//func (app *App) run() {
//	app.buildTemplate()
//	if app.Config.HttpAddr == "" {
//		app.Config.HttpAddr = "0.0.0.0"
//	}

//	addr := fmt.Sprintf("%s:%d", app.Config.HttpAddr, app.Config.HttpPort)
//	var err error

//	for {
//		if app.Config.UseFcgi {
//			l, e := net.Listen("tcp", addr)
//			if e != nil {
//				Logger.Info("Listen: %v", e)
//			}
//			err = fcgi.Serve(l, app.Handlers)
//		} else {
//			err = httpListenAndServe(addr, app.Handlers, app.Config.ReadTimeout, app.Config.WriteTimeout)
//		}
//		if err != nil {
//			Logger.Info("ListenAndServe: %v", err)
//		}
//		time.Sleep(time.Second * 2)
//	}
//}

//func httpListenAndServe(addr string, handler http.Handler, readTimeout time.Duration, writeTimeout time.Duration) error {
//	server := &http.Server{Addr: addr, Handler: handler, ReadTimeout: readTimeout, WriteTimeout: writeTimeout}
//	return server.ListenAndServe()
//}

type HttpServeListener interface {
	ListenAndServe(app *App, addr string, handler http.Handler) error
}

//Default
type DefaultHttpServeListener struct {
	Network string
}

func (hsl *DefaultHttpServeListener) ListenAndServe(app *App, addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler, ReadTimeout: app.Config.ReadTimeout, WriteTimeout: app.Config.WriteTimeout}
	if w, ok := Logger.(io.Writer); ok {
		server.ErrorLog = log.New(w, "[HTTP]", 0)
	}
	if hsl.Network != "" {
		l, err := net.Listen(hsl.Network, addr)
		if err != nil {
			return err
		}
		return server.Serve(l)
	} else {
		return server.ListenAndServe()
	}
}

//TLS
type TLSHttpServeListener struct {
	CertFile, KeyFile string
}

func (hsl *TLSHttpServeListener) ListenAndServe(app *App, addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler, ReadTimeout: app.Config.ReadTimeout, WriteTimeout: app.Config.WriteTimeout}
	if w, ok := Logger.(io.Writer); ok {
		server.ErrorLog = log.New(w, "[HTTPS]", 0)
	}
	return server.ListenAndServeTLS(hsl.CertFile, hsl.KeyFile)
}

//fcgi
type FcgiHttpServeListener struct {
	EnableStdIo bool
}

func (hsl *FcgiHttpServeListener) ListenAndServe(app *App, addr string, handler http.Handler) error {
	var err error
	var l net.Listener
	if hsl.EnableStdIo {
		if err = fcgi.Serve(nil, handler); err == nil {
			Logger.Info("Use FCGI via standard I/O")
			return nil
		} else {
			return errors.New(fmt.Sprintf("Cannot use FCGI via standard I/O, %v", err))
		}
	}
	if app.Config.HttpPort == 0 {
		if fileExists(addr) {
			os.Remove(addr)
		}
		l, err = net.Listen("unix", addr)
	} else {
		l, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return errors.New(fmt.Sprintf("Listen: %v", err))
	}
	if err = fcgi.Serve(l, handler); err != nil {
		return errors.New(fmt.Sprintf("Fcgi.Serve: %v", err))
	}
	return nil
}
