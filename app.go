package ssss

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"path"
	"strings"
	"time"
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

func (app *App) Run() {
	app.buildTemplate()
	if app.Config.HttpAddr == "" {
		app.Config.HttpAddr = "0.0.0.0"
	}

	addr := fmt.Sprintf("%s:%d", app.Config.HttpAddr, app.Config.HttpPort)
	var err error

	for {
		if app.Config.UseFcgi {
			l, e := net.Listen("tcp", addr)
			if e != nil {
				Logger.Info("Listen: %v", e)
			}
			err = fcgi.Serve(l, app.Handlers)
		} else {
			err = httpListenAndServe(addr, app.Handlers, app.Config.ReadTimeout, app.Config.WriteTimeout)
		}
		if err != nil {
			Logger.Info("ListenAndServe: %v", err)
		}
		time.Sleep(time.Second * 2)
	}
}

func httpListenAndServe(addr string, handler http.Handler, readTimeout time.Duration, writeTimeout time.Duration) error {
	server := &http.Server{Addr: addr, Handler: handler, ReadTimeout: readTimeout, WriteTimeout: writeTimeout}
	return server.ListenAndServe()
}
