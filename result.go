package ssss

import (
	"fmt"
	"strconv"
	"strings"
)

type Result struct {
	Code    int         `json:"code" xml:"code,attr"` //0为成功，其它值为错误码
	Message string      `json:"message,omitempty" xml:"message,attr,omitempty"`
	Info    interface{} `json:"info,omitempty" xml:"info,omitempty"` //具体结果数据, 只有当code为0时，才设置此属性值
}

func (r *Result) String() string {
	return "[" + strconv.Itoa(r.Code) + "] " + r.Message
}

func NewResult(code int, convertCodeToMsg bool, data ...interface{}) *Result {
	msg := ""
	if convertCodeToMsg {
		msg = ERROR_INFO_MAP[code]
	}
	if code == ERROR_CODE_OK {
		var info interface{}
		if len(data) == 1 {
			info = data[0]
		} else if len(data) > 1 {
			info = data
		}
		return &Result{Code: code, Info: info, Message: msg}
	} else {
		if len(data) > 0 {
			msg = fmt.Sprint(msg, ", ", joinMsgs(data...))
		}
		return &Result{Code: code, Message: msg}
	}
}

func NewErrorResult(code int, msgs ...interface{}) *Result {
	return NewResult(code, true, msgs...)
}

func NewSucceedResult(info interface{}) *Result {
	return &Result{Code: ERROR_CODE_OK, Info: info}
}

func joinMsgs(args ...interface{}) string {
	strs := make([]string, 0, len(args)*2)
	for _, arg := range args {
		strs = append(strs, fmt.Sprint(arg))
	}
	return strings.Join(strs, ", ")
}

func isErrorResult(err interface{}) bool {
	switch e := err.(type) {
	case *Result:
		return e.Code != ERROR_CODE_OK
	case Result:
		return e.Code != ERROR_CODE_OK
	}
	return false
}

func isResult(err interface{}) bool {
	switch err.(type) {
	case *Result, Result:
		return true
	}
	return false
}

func convertErrorResult(err interface{}) *Result {
	switch e := err.(type) {
	case *Result:
		return e
	case Result:
		return &e
	case error:
		return NewErrorResult(ERROR_CODE_RUNTIME, e.Error())
	}
	if err != nil {
		return NewErrorResult(ERROR_CODE_RUNTIME, fmt.Sprint(err))
	}
	return NewErrorResult(ERROR_CODE_RUNTIME)
}
