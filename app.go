package ssss

import (
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
	Logger           Logger
}

func NewApp() *App {
	app := &App{
		Handlers:         NewControllerRegister(nil),
		TemplateRegister: NewTemplateRegister(nil),
		Config:           newConfig(),
		StaticDirs:       make(map[string]string),
		Logger:           &defaultLogger{}}

	app.Handlers.app = app
	app.TemplateRegister.app = app
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
			app.Logger.Error("%v", err)
		}
	}
}

func (app *App) Prepare() {
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

	if err = s.ListenAndServe(app); err != nil {
		app.Logger.Critical("%v.ListenAndServe: %v", reflect.TypeOf(s), err)
	}

}
