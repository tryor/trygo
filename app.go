package trygo

import (
	"net"
	"net/http"
	"path"
	"reflect"
	"strings"
)

type App struct {
	Handlers         *ControllerRegister
	Config           *config
	StaticDirs       map[string]string
	TemplateRegister *TemplateRegister
	Logger           LoggerInterface
	Statinfo         *statinfo
	//filter net.Listener
	FilterListener func(app *App, l net.Listener) net.Listener
	//filter http.Handler
	FilterHandler func(app *App, h http.Handler) http.Handler

	prepared bool
}

func NewApp() *App {
	app := &App{
		Handlers:         NewControllerRegister(nil),
		TemplateRegister: NewTemplateRegister(nil),
		Config:           newConfig(),
		StaticDirs:       make(map[string]string),
		Logger:           Logger,
	}

	app.FilterListener = func(app *App, l net.Listener) net.Listener { return l }
	app.FilterHandler = func(app *App, h http.Handler) http.Handler { return h }
	app.Statinfo = newStatinfo(app)
	app.Handlers.app = app
	app.TemplateRegister.app = app
	return app
}

//method - http method, GET,POST,PUT,HEAD,DELETE,PATCH,OPTIONS,*
//pattern - URL path or regexp pattern
//name - method on the container
//tags - function parameter tag info, see struct tag
func (app *App) Register(method string, pattern string, c ControllerInterface, name string, tags ...string) iRouter {
	funcname, params := parseMethod(name)
	return app.Handlers.Add(method, pattern, c, funcname, params, tags)
}

func (app *App) RegisterHandler(pattern string, h http.Handler) iRouter {
	return app.Handlers.AddHandler(pattern, h)
}

func (app *App) RegisterFunc(methods, pattern string, f HandlerFunc) iRouter {
	return app.Handlers.AddMethod(methods, pattern, f)
}

func (app *App) RegisterRESTful(pattern string, c ControllerInterface) iRouter {
	//app.Register("*", pattern, c, "")
	if !isPattern(pattern) {
		pattern = path.Join(pattern, "(?P<id>[^/]+)$")
	}
	return app.Register("*", pattern, c, "")
}

func (app *App) Get(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Get(pattern, f)
}

func (app *App) Post(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Post(pattern, f)
}

func (app *App) Put(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Put(pattern, f)
}

func (app *App) Delete(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Delete(pattern, f)
}

func (app *App) Head(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Head(pattern, f)
}

func (app *App) Patch(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Patch(pattern, f)
}

func (app *App) Options(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Options(pattern, f)
}

func (app *App) Any(pattern string, f HandlerFunc) iRouter {
	return app.Handlers.Any(pattern, f)
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
			app.Logger.Error("%v", err)
		}
	}
}

func (app *App) Prepare() {
	if app.prepared {
		return
	}
	app.prepared = true
	app.buildTemplate()
}

func (app *App) Run(server ...Server) {
	var err error
	var s Server
	if len(server) > 0 {
		s = server[0]
	} else {
		s = &HttpServer{}
	}
	app.Prepare()
	if err = s.ListenAndServe(app); err != nil {
		app.Logger.Critical("%v.ListenAndServe: %v", reflect.TypeOf(s), err)
	}

}
