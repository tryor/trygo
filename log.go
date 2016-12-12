package trygo

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Logger interface {
	Debug(arg0 interface{}, args ...interface{})
	Info(arg0 interface{}, args ...interface{})
	Warn(arg0 interface{}, args ...interface{}) error
	Error(arg0 interface{}, args ...interface{}) error
	Critical(arg0 interface{}, args ...interface{}) error
}

var logger Logger

func init() {
	SetLogger(&defaultLogger{})
}

func SetLogger(l Logger) {
	logger = l
}

type defaultLogger struct{}

func (l *defaultLogger) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if p[len(p)-1] == '\n' {
		p = p[0 : len(p)-1]
	}
	fmt.Printf("%s [LOG] %s %s\n", formatNow(), determine(5), string(p))
	return len(p), nil
}

func (l *defaultLogger) Printf(format string, args ...interface{}) {
	fmt.Printf("%s [LOG] %s %s\n", formatNow(), determine(3), fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Debug(arg0 interface{}, args ...interface{}) {
	switch f := arg0.(type) {
	case string:
		l.log("DBG", f, args...)
	default:
		l.log("DBG", fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (l *defaultLogger) Info(arg0 interface{}, args ...interface{}) {
	switch f := arg0.(type) {
	case string:
		l.log("INF", f, args...)
	default:
		l.log("INF", fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (l *defaultLogger) Warn(arg0 interface{}, args ...interface{}) error {
	switch f := arg0.(type) {
	case string:
		return errors.New(l.log("WRN", f, args...))
	default:
		return errors.New(l.log("WRN", fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...))
	}
}

func (l *defaultLogger) Error(arg0 interface{}, args ...interface{}) error {
	switch f := arg0.(type) {
	case string:
		return errors.New(l.log("ERR", f, args...))
	default:
		return errors.New(l.log("ERR", fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...))
	}
}

func (l *defaultLogger) Critical(arg0 interface{}, args ...interface{}) error {
	switch f := arg0.(type) {
	case string:
		return errors.New(l.log("CRT", f, args...))
	default:
		return errors.New(l.log("CRT", fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...))
	}
}

func (l *defaultLogger) log(lvl string, format string, args ...interface{}) (msg string) {
	msg = fmt.Sprintf(format, args...)
	fmt.Printf("%s [%s] %s %s\n", formatNow(), lvl, determine(3), msg)
	return
}

func formatNow() string {
	return time.Now().Format("2006-01-02 15:04:05.999")
}

func determine(skip int) string {
	pc, file, lineno, ok := runtime.Caller(skip)
	src := ""
	if ok {
		name := runtime.FuncForPC(pc).Name()
		nameitems := strings.Split(name, ".")
		if len(nameitems) > 2 {
			nameitems = nameitems[len(nameitems)-2:]
		}
		name = strings.Join(nameitems, ".")

		pathitems := strings.Split(file, "/")
		if len(pathitems) > 2 {
			pathitems = pathitems[len(pathitems)-2:]
		}
		file = strings.Join(pathitems, "/")
		src = fmt.Sprintf("%s:%d(%s)", file, lineno, name)
	}
	return src
}
