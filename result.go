package ssss

import (
	"fmt"
	"strconv"
	"strings"
)

//是否自动构建消息
var AutoBuildMessage bool = true

type Result struct {
	Code    int         `json:"code" xml:"code"` //0为成功，其它值为错误码
	Message string      `json:"message,omitempty" xml:"message,omitempty"`
	Info    interface{} `json:"info,omitempty" xml:"info,omitempty"` //具体结果数据, 只有当code为0时，才设置此属性值
}

func (r *Result) String() string {
	return "[" + strconv.Itoa(r.Code) + "]" + r.Message
}

func NewErrorResult(code int, msgs ...interface{}) *Result {
	msg := ""
	if AutoBuildMessage {
		msg = ERROR_INFO_MAP[code]
		if len(msgs) > 0 {
			msg = fmt.Sprint(msg, ",", joinMsgs(msgs...))
		}
	}
	return &Result{Code: code, Message: msg}
}

func NewSucceedResult(info interface{}) *Result {
	return &Result{Code: 0, Info: info}
}

func joinMsgs(args ...interface{}) string {
	strs := make([]string, 0, len(args)*2)
	for _, arg := range args {
		strs = append(strs, fmt.Sprint(arg))
	}
	return strings.Join(strs, ",")
}

//将错误转换为Result
func ConvertErrorResult(err interface{}) *Result {
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
	return NewErrorResult(ERROR_CODE_RUNTIME, ERROR_INFO_MAP[ERROR_CODE_RUNTIME])
}
