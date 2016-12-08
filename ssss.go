package ssss

import (
	"net/http"
	"os"
)

const VERSION = "0.1.0"

var (
	DefaultApp *App
	AppPath    string
)

func init() {
	DefaultApp = NewApp()
	AppPath, _ = os.Getwd()
}

func Register(method string, pattern string, c IController, name string, tags ...string) *App {
	DefaultApp.Register(method, pattern, c, name, tags...)
	return DefaultApp
}

func RegisterHandler(pattern string, h http.Handler) *App {
	DefaultApp.RegisterHandler(pattern, h)
	return DefaultApp
}

func RegisterRESTful(pattern string, c IController) *App {
	DefaultApp.RegisterRESTful(pattern, c)
	return DefaultApp
}

func RegisterFunc(methods, pattern string, f HandlerFunc) *App {
	DefaultApp.RegisterFunc(methods, pattern, f)
	return DefaultApp
}

func Get(pattern string, f HandlerFunc) *App {
	DefaultApp.Get(pattern, f)
	return DefaultApp
}

func Post(pattern string, f HandlerFunc) *App {
	DefaultApp.Post(pattern, f)
	return DefaultApp
}

func Put(pattern string, f HandlerFunc) *App {
	DefaultApp.Put(pattern, f)
	return DefaultApp
}

func Delete(pattern string, f HandlerFunc) *App {
	DefaultApp.Delete(pattern, f)
	return DefaultApp
}

func Head(pattern string, f HandlerFunc) *App {
	DefaultApp.Head(pattern, f)
	return DefaultApp
}

func Patch(pattern string, f HandlerFunc) *App {
	DefaultApp.Patch(pattern, f)
	return DefaultApp
}

func Options(pattern string, f HandlerFunc) *App {
	DefaultApp.Handlers.Options(pattern, f)
	return DefaultApp
}

func Any(pattern string, f HandlerFunc) *App {
	DefaultApp.Any(pattern, f)
	return DefaultApp
}

func SetStaticPath(url string, path string) *App {
	DefaultApp.SetStaticPath(url, path)
	return DefaultApp
}

func SetViewsPath(path string) *App {
	DefaultApp.SetViewsPath(path)
	return DefaultApp
}

func AddTemplateExt(ext string) {
	DefaultApp.TemplateRegister.AddTemplateExt(ext)
}

func AddTemplateFunc(key string, funname interface{}) error {
	return DefaultApp.TemplateRegister.AddFuncMap(key, funname)
}

func Run(listener ...HttpServeListener) {
	DefaultApp.Run(listener...)
}
