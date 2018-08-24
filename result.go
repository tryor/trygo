package trygo

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
	if r.Code == ERROR_CODE_OK {
		return fmt.Sprint(r.Info)
	} else {
		return "[" + strconv.Itoa(r.Code) + "] " + r.Message
	}
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
			if msg == "" {
				msg = fmt.Sprint(joinMsgs(data...))
			} else {
				msg = fmt.Sprint(msg, ", ", joinMsgs(data...))
			}

		}
		return &Result{Code: code, Message: msg}
	}
}

func NewErrorResult(code int, msgs ...interface{}) *Result {
	if len(msgs) == 0 {
		return NewResult(code, true)
	} else {
		return NewResult(code, false, msgs...)
	}
}

func NewSucceedResult(info interface{}) *Result {
	return &Result{Code: ERROR_CODE_OK, Info: info}
}

func joinMsgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	if len(args) == 1 {
		return fmt.Sprint(args[0])
	}

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

func convertErrorResult(err interface{}, code ...int) *Result {
	switch e := err.(type) {
	case *Result:
		return e
	case Result:
		return &e
		//	case error:
		//		return NewErrorResult(ERROR_CODE_RUNTIME, e.Error())
	}

	var c int
	if len(code) > 0 && code[0] != ERROR_CODE_OK {
		c = code[0]
	} else {
		c = ERROR_CODE_RUNTIME
	}
	if err != nil {
		return NewErrorResult(c, fmt.Sprint(err))
	}
	return NewErrorResult(c)
}
