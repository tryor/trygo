package trygo

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

type Server interface {
	ListenAndServe(app *App) error
}

//Default
type HttpServer struct {
	http.Server
	Network string
}

func (hsl *HttpServer) ListenAndServe(app *App) error {
	app.Prepare()
	hsl.ReadTimeout = app.Config.Listen.ReadTimeout
	hsl.WriteTimeout = app.Config.Listen.WriteTimeout
	hsl.Addr = app.Config.Listen.Addr
	hsl.Handler = app.FilterHandler(app, DefaultFilterHandler(app, app.Handlers))

	if w, ok := app.Logger.(io.Writer); ok {
		hsl.ErrorLog = log.New(w, "[HTTP]", 0)
	}
	if hsl.Network == "" {
		hsl.Network = "tcp"
	}
	l, err := net.Listen(hsl.Network, hsl.Addr)
	if err != nil {
		return err
	}
	return hsl.Serve(app.FilterListener(app, DefaultFilterListener(app, l)))
}

//TLS
type TLSHttpServer struct {
	http.Server
	CertFile, KeyFile string
}

func (hsl *TLSHttpServer) ListenAndServe(app *App) error {
	app.Prepare()
	hsl.ReadTimeout = app.Config.Listen.ReadTimeout
	hsl.WriteTimeout = app.Config.Listen.WriteTimeout
	hsl.Addr = app.Config.Listen.Addr
	hsl.Handler = app.FilterHandler(app, DefaultFilterHandler(app, app.Handlers))

	if w, ok := app.Logger.(io.Writer); ok {
		hsl.ErrorLog = log.New(w, "[HTTPS]", 0)
	}

	config, err := hsl.tlsConfig()
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", hsl.Addr)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(app.FilterListener(app, DefaultFilterListener(app, l)), config)
	return hsl.Serve(tlsListener)
}

func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func (hsl *TLSHttpServer) tlsConfig() (*tls.Config, error) {
	config := &tls.Config{}
	if !strSliceContains(config.NextProtos, "http/1.1") {
		config.NextProtos = append(config.NextProtos, "http/1.1")
	}
	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert || hsl.CertFile != "" || hsl.KeyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(hsl.CertFile, hsl.KeyFile)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

//fcgi
type FcgiHttpServer struct {
	Network     string
	EnableStdIo bool
}

func (hsl *FcgiHttpServer) ListenAndServe(app *App) error {
	app.Prepare()
	var err error
	var l net.Listener
	handler := app.FilterHandler(app, DefaultFilterHandler(app, app.Handlers))
	addr := app.Config.Listen.Addr

	if hsl.EnableStdIo {
		if err = fcgi.Serve(nil, handler); err == nil {
			app.Logger.Info("Use FCGI via standard I/O")
			return nil
		} else {
			return errors.New(fmt.Sprintf("Cannot use FCGI via standard I/O, %v", err))
		}
	}
	if hsl.Network == "unix" {
		if fileExists(addr) {
			os.Remove(addr)
		}
		l, err = net.Listen("unix", addr)
	} else {
		network := hsl.Network
		if network == "" {
			network = "tcp"
		}
		l, err = net.Listen(network, addr)
	}
	if err != nil {
		return errors.New(fmt.Sprintf("Listen: %v", err))
	}
	if err = fcgi.Serve(app.FilterListener(app, DefaultFilterListener(app, l)), handler); err != nil {
		return errors.New(fmt.Sprintf("Fcgi.Serve: %v", err))
	}
	return nil
}
