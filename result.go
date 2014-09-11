package ssss

import (
	"fmt"
)

import ()

type Result struct {
	Code    int         `json:"code" xml:"code"` //0为成功，其它值为错误码
	Message string      `json:"message,omitempty" xml:"message,omitempty"`
	Info    interface{} `json:"info,omitempty" xml:"info,omitempty"` //具体结果数据, 只有当code为0时，才设置此属性值
}

func errorResult(code int, msgs ...interface{}) *Result {
	if len(msgs) > 0 {
		return &Result{Code: code, Message: fmt.Sprint(msgs...)}
	}
	return &Result{Code: code}
}

func succeedResult(info interface{}) *Result {
	return &Result{Code: 0, Info: info}
}

//将错误转换为Result
func convertErrorResult(err interface{}) *Result {
	switch e := err.(type) {
	case *Result:
		return e
	case Result:
		return &e
	case error:
		return errorResult(ERROR_CODE_RUNTIME, e.Error())
	}
	if err != nil {
		return errorResult(ERROR_CODE_RUNTIME, fmt.Sprint(err))
	}
	return errorResult(ERROR_CODE_RUNTIME, "运行时异常")
}
