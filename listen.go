package ssss

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
	"sync"
	"time"
)

type HttpServeListener interface {
	ListenAndServe(app *App, addr string, handler http.Handler) error
}

//Default
type DefaultHttpServeListener struct {
	http.Server
	Network string
}

func (hsl *DefaultHttpServeListener) ListenAndServe(app *App, addr string, handler http.Handler) error {
	//server := &http.Server{ReadTimeout: app.Config.Listen.ReadTimeout, WriteTimeout: app.Config.Listen.WriteTimeout}
	hsl.ReadTimeout = app.Config.Listen.ReadTimeout
	hsl.WriteTimeout = app.Config.Listen.WriteTimeout
	hsl.Addr = addr
	hsl.Handler = FilterHandler(app, handler)

	if w, ok := app.Logger.(io.Writer); ok {
		hsl.ErrorLog = log.New(w, "[HTTP]", 0)
	}
	if hsl.Network == "" {
		hsl.Network = "tcp"
	}
	l, err := net.Listen(hsl.Network, addr)
	if err != nil {
		return err
	}
	return hsl.Serve(FilterListener(app, l))
}

//TLS
type TLSHttpServeListener struct {
	http.Server
	CertFile, KeyFile string
}

func (hsl *TLSHttpServeListener) ListenAndServe(app *App, addr string, handler http.Handler) error {
	//	server := &http.Server{ReadTimeout: app.Config.Listen.ReadTimeout, WriteTimeout: app.Config.Listen.WriteTimeout}
	//	server.Addr = addr
	//	server.Handler = FilterHandler(app, handler)
	hsl.ReadTimeout = app.Config.Listen.ReadTimeout
	hsl.WriteTimeout = app.Config.Listen.WriteTimeout
	hsl.Addr = addr
	hsl.Handler = FilterHandler(app, handler)

	if w, ok := app.Logger.(io.Writer); ok {
		hsl.ErrorLog = log.New(w, "[HTTPS]", 0)
	}

	config, err := hsl.tlsConfig()
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(FilterListener(app, l), config)
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

func (hsl *TLSHttpServeListener) tlsConfig() (*tls.Config, error) {
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
type FcgiHttpServeListener struct {
	Network     string
	EnableStdIo bool
}

func (hsl *FcgiHttpServeListener) ListenAndServe(app *App, addr string, handler http.Handler) error {
	var err error
	var l net.Listener
	handler = FilterHandler(app, handler)

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
	if err = fcgi.Serve(FilterListener(app, l), handler); err != nil {
		return errors.New(fmt.Sprintf("Fcgi.Serve: %v", err))
	}
	return nil
}

func FilterListener(app *App, l net.Listener) net.Listener {
	l = TcpKeepAliveListener(l, app.Config.Listen.MaxKeepaliveDuration)
	l = LimitListener(l, app.Config.Listen.Concurrency)
	return l
}

func TcpKeepAliveListener(l net.Listener, keepalivePeriod time.Duration) net.Listener {
	if tc, ok := l.(*net.TCPListener); ok {
		return &tcpKeepAliveListener{tc, keepalivePeriod}
	}
	DefaultApp.Logger.Warn("Listen: Listener not is *net.TCPListener, %v", l.Addr())
	return l
}

type tcpKeepAliveListener struct {
	*net.TCPListener
	keepalivePeriod time.Duration
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.TCPListener.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	if ln.keepalivePeriod > 0 {
		tc.SetKeepAlivePeriod(ln.keepalivePeriod)
	} else {
		tc.SetKeepAlivePeriod(time.Minute * 3)
	}
	return tc, nil
}

func LimitListener(l net.Listener, n int) net.Listener {
	return &limitListener{l, make(chan struct{}, n)}
}

type limitListener struct {
	net.Listener
	sem chan struct{}
}

func (l *limitListener) acquire() {
	l.sem <- struct{}{}
}
func (l *limitListener) release() {
	<-l.sem
}

func (l *limitListener) Accept() (net.Conn, error) {
	l.acquire()
	c, err := l.Listener.Accept()
	if err != nil {
		l.release()
		return nil, err
	}
	return &limitListenerConn{Conn: c, release: l.release}, nil
}

type limitListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

func (l *limitListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(l.release)
	return err
}
