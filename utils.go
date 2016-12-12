package trygo

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	maxValueLength   = 4096
	maxHeaderLines   = 1024
	chunkSize        = 4 << 10  // 4 KB chunks
	defaultMaxMemory = 32 << 20 // 32 MB
)

func toString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case *string:
		return *s
	default:
		return fmt.Sprint(v)
	}

}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func parseForm(r *http.Request, multipart bool) (url.Values, error) {
	if multipart {
		err := r.ParseMultipartForm(defaultMaxMemory)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			return nil, err
		}
	}
	return r.Form, nil
}

func isPattern(path string) bool {
	return strings.ContainsAny(path, "^.()[]|*+?{}\\,<>:=!$")

}

func newValueOf(kind reflect.Kind) (*reflect.Value, error) {
	var v reflect.Value
	switch kind {
	case reflect.String:
		v = reflect.New(reflect.TypeOf(""))
	case reflect.Bool:
		v = reflect.New(reflect.TypeOf(false))
	case reflect.Int:
		v = reflect.New(reflect.TypeOf(int(0)))
	case reflect.Int8:
		v = reflect.New(reflect.TypeOf(int8(0)))
	case reflect.Int16:
		v = reflect.New(reflect.TypeOf(int16(0)))
	case reflect.Int32:
		v = reflect.New(reflect.TypeOf(int32(0)))
	case reflect.Int64:
		v = reflect.New(reflect.TypeOf(int64(0)))

	case reflect.Uint:
		v = reflect.New(reflect.TypeOf(uint(0)))
	case reflect.Uint8:
		v = reflect.New(reflect.TypeOf(uint8(0)))
	case reflect.Uint16:
		v = reflect.New(reflect.TypeOf(uint16(0)))
	case reflect.Uint32:
		v = reflect.New(reflect.TypeOf(uint32(0)))
	case reflect.Uint64:
		v = reflect.New(reflect.TypeOf(uint64(0)))
	case reflect.Float32:
		v = reflect.New(reflect.TypeOf(float32(0.0)))
	case reflect.Float64:
		v = reflect.New(reflect.TypeOf(float64(0.0)))
	default:
		return nil, errors.New("the type (" + kind.String() + ") is not supported")
	}

	v = reflect.Indirect(v)

	return &v, nil
}

//解析string类型值到指定其它类型值
//srcValue - 原值
//destValue - 在destValue中返回与之相同类型的值, 如果不指定destValue，将自动创建    //destType - 转换目标类型
//@return 与destType相同类型的值
func parseValue(srcValue string, destTypeKind reflect.Kind, destValue *reflect.Value) (*reflect.Value, error) {
	var err error
	if destValue == nil {
		destValue, err = newValueOf(destTypeKind)
		if err != nil {
			return nil, err
		}
	}

	var v interface{}
	switch destTypeKind {
	case reflect.String:
		destValue.SetString(srcValue)
	case reflect.Bool:
		v, err = strconv.ParseBool(srcValue)
		if err == nil {
			destValue.SetBool(v.(bool))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err = strconv.ParseInt(srcValue, 10, 64)
		if err == nil {
			destValue.SetInt(v.(int64))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err = strconv.ParseUint(srcValue, 10, 64)
		if err == nil {
			destValue.SetUint(v.(uint64))
		}
	case reflect.Float32:
		v, err = strconv.ParseFloat(srcValue, 32)
		if err == nil {
			destValue.SetFloat(v.(float64))
		}
	case reflect.Float64:
		v, err = strconv.ParseFloat(srcValue, 64)
		if err == nil {
			destValue.SetFloat(v.(float64))
		}
	default:
		err = errors.New("the type (" + destTypeKind.String() + ") is not supported")
	}

	return destValue, err
}

const (
	tcUnknown = iota
	tcSigned
	tcUnsigned
	tcString
	tcBool
	tcFloat
)

func typeClassify(typ reflect.Type) int8 {
	switch typ.Kind() {
	case reflect.String:
		return tcString
	case reflect.Bool:
		return tcBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return tcSigned
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return tcUnsigned
	case reflect.Float32, reflect.Float64:
		return tcFloat
	default:
		return tcUnknown
	}
}

func toInt64(v interface{}) int64 {
	switch vi := v.(type) {
	case int:
		return int64(vi)
	case int64:
		return vi
	case int32:
		return int64(vi)
	case int16:
		return int64(vi)
	case int8:
		return int64(vi)

	case uint:
		return int64(vi)
	case uint64:
		return int64(vi)
	case uint32:
		return int64(vi)
	case uint16:
		return int64(vi)
	case uint8:
		return int64(vi)

	case float32:
		return int64(vi)
	case float64:
		return int64(vi)

	case bool:
		if vi {
			return 1
		} else {
			return 0
		}

	case string:
		vi64, err := strconv.ParseInt(vi, 10, 64)
		if err != nil {
			panic(err)
		}
		return vi64
	default:
		panic(errors.New("unknown data type"))
	}
}

func toUint64(v interface{}) uint64 {
	switch vi := v.(type) {
	case uint:
		return uint64(vi)
	case uint64:
		return vi
	case uint32:
		return uint64(vi)
	case uint16:
		return uint64(vi)
	case uint8:
		return uint64(vi)

	case int:
		return uint64(vi)
	case int64:
		return uint64(vi)
	case int32:
		return uint64(vi)
	case int16:
		return uint64(vi)
	case int8:
		return uint64(vi)

	case float32:
		return uint64(vi)
	case float64:
		return uint64(vi)

	case bool:
		if vi {
			return 1
		} else {
			return 0
		}
	case string:
		vu64, err := strconv.ParseUint(vi, 10, 64)
		if err != nil {
			panic(err)
		}
		return vu64
	default:
		panic(errors.New("unknown data type"))
	}
}

func toFloat64(v interface{}) float64 {
	switch vi := v.(type) {
	case int:
		return float64(vi)
	case int64:
		return float64(vi)
	case int32:
		return float64(vi)
	case int16:
		return float64(vi)
	case int8:
		return float64(vi)

	case uint:
		return float64(vi)
	case uint64:
		return float64(vi)
	case uint32:
		return float64(vi)
	case uint16:
		return float64(vi)
	case uint8:
		return float64(vi)

	case float32:
		return float64(vi)
	case float64:
		return vi

	case bool:
		if vi {
			return 1.0
		} else {
			return 0.0
		}

	case string:
		vi64, err := strconv.ParseFloat(vi, 64)
		if err != nil {
			panic(err)
		}
		return vi64
	default:
		panic(errors.New("unknown data type"))
	}
}

//从方法中分离出方法名称和参数
//Login(account, pwd string)
func parseMethod(method string) (name string, params []string) {
	pairs := strings.SplitN(method, "(", 2)
	name = strings.TrimSpace(pairs[0])
	if len(pairs) > 1 {
		paramsstr := strings.TrimSpace(strings.Replace(pairs[1], ")", "", -1))
		if len(paramsstr) > 0 {
			params = strings.Split(paramsstr, ",")
		}
	}
	return
}

func getContentType(typ string) string {
	ext := typ
	if !strings.HasPrefix(typ, ".") {
		ext = "." + typ
	}

	ct := mimemaps[ext]
	if ct != "" {
		return ct
	}

	ct = mime.TypeByExtension(ext)
	if ct == "" {
		return typ
	}
	return ct
}

func toContentType(format string) string {
	switch format {
	case "json":
		return "application/json; charset=utf-8"
	case "xml":
		return "text/xml; charset=utf-8"
	case "txt":
		return "text/plain; charset=utf-8"
	case "html":
		return "text/html; charset=utf-8"
	}
	return getContentType(format)
}

func buildLoginfo(r *http.Request, args ...interface{}) string {
	return fmt.Sprintf("%s \"%s\"<->\"%s\": %s", r.URL.Path, r.Host, r.RemoteAddr, fmt.Sprint(args...))
}
