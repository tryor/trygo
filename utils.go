package ssss

import (
	"errors"
	"strconv"
	"strings"
)

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
