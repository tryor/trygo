package ssss

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type ILogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Critical(format string, args ...interface{})
}

var Logger ILogger

func SetLogger(logger ILogger) {
	Logger = logger
}

func init() {
	if Logger == nil {
		Logger = &defaultLogger{}
	}
}

type defaultLogger struct{}

func (this *defaultLogger) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if p[len(p)-1] == '\n' {
		p = p[0 : len(p)-1]
	}
	fmt.Printf("%v [LOG] %v %v\n", formatNow(), determine(5), string(p))
	return len(p), nil
}

func (this *defaultLogger) Debug(format string, args ...interface{}) {
	fmt.Printf("%v [DBG] %v %v\n", formatNow(), determine(2), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Info(format string, args ...interface{}) {
	fmt.Printf("%v [INF] %v %v\n", formatNow(), determine(2), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Warn(format string, args ...interface{}) {
	fmt.Printf("%v [WRN] %v %v\n", formatNow(), determine(2), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Error(format string, args ...interface{}) {
	fmt.Printf("%v [ERR] %v %v\n", formatNow(), determine(2), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Critical(format string, args ...interface{}) {
	fmt.Printf("%v [CRT] %v %v\n", formatNow(), determine(2), fmt.Sprintf(format, args...))
}

func formatNow() string {
	return time.Now().Format("2006-01-02 15:04:05.999")
}

func determine(skip int) string {
	pc, file, lineno, ok := runtime.Caller(skip)
	src := ""
	if ok {
		name := runtime.FuncForPC(pc).Name()
		pathitems := strings.Split(file, "/")
		if len(pathitems) > 2 {
			pathitems = pathitems[len(pathitems)-2:]
		}
		file = strings.Join(pathitems, "/")
		src = fmt.Sprintf("%s:%d(%s)", file, lineno, name)
	}
	return src
}
