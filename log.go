package ssss

import (
	"fmt"
	"runtime"
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

func (this *defaultLogger) Debug(format string, args ...interface{}) {
	fmt.Printf("%v [DBG] %v %v\n", formatNow(), determine(), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Info(format string, args ...interface{}) {
	fmt.Printf("%v [INF] %v %v\n", formatNow(), determine(), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Warn(format string, args ...interface{}) {
	fmt.Printf("%v [WRN] %v %v\n", formatNow(), determine(), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Error(format string, args ...interface{}) {
	fmt.Printf("%v [ERR] %v %v\n", formatNow(), determine(), fmt.Sprintf(format, args...))
}

func (this *defaultLogger) Critical(format string, args ...interface{}) {
	fmt.Printf("%v [CRT] %v %v\n", formatNow(), determine(), fmt.Sprintf(format, args...))
}

func formatNow() string {
	return time.Now().Format("2006-01-02 15:04:05.999")
}

func determine() string {
	// Determine caller func
	_, file, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		//src = fmt.Sprintf("%v/%s:%d", file, runtime.FuncForPC(pc).Name(), lineno)
		src = fmt.Sprintf("%s:%d", file, lineno)
	}
	return src
}
