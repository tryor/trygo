package ssss

import (
	"net"
	"sync"
	"time"
)

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
