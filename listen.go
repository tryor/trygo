package trygo

import (
	"errors"
	"net"
	"sync"
	"time"
)

func DefaultFilterListener(app *App, l net.Listener) net.Listener {
	l = TcpKeepAliveListener(l, time.Minute*3)
	l = LimitListener(l, app)
	//if app.Config.Listen.MaxKeepaliveDuration > 0 {
	//	l = LimitKeepaliveDurationListener(l, app.Config.Listen.MaxKeepaliveDuration)
	//}
	return l
}

func LimitKeepaliveDurationListener(l net.Listener, maxKeepaliveDuration time.Duration) net.Listener {
	return &limitKeepaliveDurationListener{l, maxKeepaliveDuration}
}

var ErrKeepaliveTimeout = errors.New("exceeded MaxKeepaliveDuration")

type limitKeepaliveDurationListener struct {
	net.Listener
	maxKeepaliveDuration time.Duration
}

func (l limitKeepaliveDurationListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &limitConnKeepaliveDuration{Conn: c, expire: time.Now().Add(l.maxKeepaliveDuration)}, nil
}

type limitConnKeepaliveDuration struct {
	net.Conn
	expire time.Time
}

func (c *limitConnKeepaliveDuration) Read(b []byte) (n int, err error) {
	if time.Now().After(c.expire) {
		c.Close()
		return 0, ErrKeepaliveTimeout
	}
	return c.Conn.Read(b)
}

func TcpKeepAliveListener(l net.Listener, keepalivePeriod time.Duration) net.Listener {
	if tc, ok := l.(*net.TCPListener); ok {
		return &tcpKeepAliveListener{tc, keepalivePeriod}
	}
	Logger.Warn("Listen: Listener not is *net.TCPListener, %v", l.Addr())
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

func LimitListener(l net.Listener, app *App) net.Listener {
	if app.Config.StatinfoEnable {
		return &limitAndStatinfoListener{limitListener: limitListener{l, make(chan struct{}, app.Config.Listen.Concurrency)}, statinfo: app.Statinfo}
	} else {
		return &limitListener{l, make(chan struct{}, app.Config.Listen.Concurrency)}
	}

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

type limitAndStatinfoListener struct {
	limitListener
	statinfo *statinfo
}

func (l *limitAndStatinfoListener) Accept() (net.Conn, error) {
	l.acquire()
	c, err := l.Listener.Accept()
	if err != nil {
		l.release()
		return nil, err
	}
	return &limitListenerConn{Conn: c, release: l.release}, nil
}

func (l *limitAndStatinfoListener) acquire() {
	l.sem <- struct{}{}
	l.statinfo.incConcurrentConns()
}
func (l *limitAndStatinfoListener) release() {
	l.statinfo.decConcurrentConns()
	<-l.sem
}
