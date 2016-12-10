package trygo

import (
	"errors"
	"net/http"
)

func FilterHandler(app *App, h http.Handler) http.Handler {
	h = BodyLimitHandler(app, h)
	return h
}

func BodyLimitHandler(app *App, handler http.Handler) http.Handler {
	return &bodyLimitHandler{app, handler}
}

type bodyLimitHandler struct {
	app     *App
	handler http.Handler
}

var ErrBodyTooLarge = errors.New("http: request body too large")

func (h *bodyLimitHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.ContentLength > h.app.Config.MaxRequestBodySize {
		h.app.Logger.Info("%s", buildLoginfo(r, ErrBodyTooLarge))
		Error(rw, ErrBodyTooLarge.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	if r.Body != nil {
		r.Body = http.MaxBytesReader(rw, r.Body, h.app.Config.MaxRequestBodySize)
	}
	h.handler.ServeHTTP(rw, r)
}
