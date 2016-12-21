package trygo

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

func Register(method string, pattern string, c ControllerInterface, name string, tags ...string) iRouter {
	return DefaultApp.Register(method, pattern, c, name, tags...)
}

func RegisterHandler(pattern string, h http.Handler) iRouter {
	return DefaultApp.RegisterHandler(pattern, h)
}

func RegisterRESTful(pattern string, c ControllerInterface) iRouter {
	return DefaultApp.RegisterRESTful(pattern, c)
}

func RegisterFunc(methods, pattern string, f HandlerFunc) iRouter {
	return DefaultApp.RegisterFunc(methods, pattern, f)
}

func Get(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Get(pattern, f)
}

func Post(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Post(pattern, f)
}

func Put(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Put(pattern, f)
}

func Delete(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Delete(pattern, f)
}

func Head(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Head(pattern, f)
}

func Patch(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Patch(pattern, f)
}

func Options(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Handlers.Options(pattern, f)
}

func Any(pattern string, f HandlerFunc) iRouter {
	return DefaultApp.Any(pattern, f)
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

func Run(server ...Server) {
	DefaultApp.Run(server...)
}
