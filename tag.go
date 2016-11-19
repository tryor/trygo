package ssss

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

//limit:20,scope:[1 2 3],default:1,require,pattern:regex
type tagInfo struct {
	Limit   limitTag
	Require bool
	Scope   scopeTag
	Default defaultTag
	Pattern patternTag
}

func (this *tagInfo) String() string {
	return fmt.Sprintf("(limit:%v, require:%v, scope:%v, default:%v, pattern:%v)", this.Limit, this.Require, this.Scope, this.Default, this.Pattern)
}

type tagItem struct {
	Exist bool
}

type limitTag struct {
	tagItem
	Value int
}

type scopeTag struct {
	tagItem
	Items scopeItems
	expr  string
}

func (this scopeTag) String() string {
	return this.expr
}

type defaultTag struct {
	tagItem
	Value reflect.Value
}

type patternTag struct {
	tagItem
	Regexp *regexp.Regexp
}

//pnames key is parmeter name, value is parmeter type
func parseTags(pnames map[string]reflect.Type, tags []string, formDomainModel bool) (formatedTags map[string]*tagInfo) {
	if len(pnames) == 0 {
		return
	}
	formatedTags = make(map[string]*tagInfo)
	alltags := reflect.StructTag(strings.Join(tags, " "))
	for pname, typ := range pnames {
		parseTag(pname, typ, alltags.Get(pname), formDomainModel, formatedTags)
	}
	//fmt.Println("formatedTags:", formatedTags)
	return
}

func parseTag(pname string, ptype reflect.Type, tag string, formDomainModel bool, formatedTags map[string]*tagInfo) {
	if ptype.Kind() == reflect.Struct {
		for i := 0; i < ptype.NumField(); i++ {
			stag := tag
			f := ptype.Field(i)
			var attrName string
			if f.Anonymous {
				attrName = pname
			} else {
				paramTag := f.Tag.Get("field")
				paramTags := strings.SplitN(paramTag, ",", 2)
				if len(paramTag) > 0 {
					attrName = strings.TrimSpace(paramTags[0])
					if len(paramTag) > 1 {
						stag = paramTags[1]
					}
				}
				if len(attrName) == 0 {
					attrName = strings.ToLower(f.Name[0:1]) + f.Name[1:]
				}
				if formDomainModel {
					attrName = pname + "." + attrName
				}
			}
			parseTag(attrName, f.Type, stag, formDomainModel, formatedTags)
		}

	} else {
		if len(tag) > 0 {
			if taginfo := parseTagItems(ptype, tag); taginfo != nil {
				formatedTags[pname] = taginfo
			}
		}
	}

}

//limit:20,scope:[1 2 3],default:1,require,pattern:regex
func parseTagItems(typ reflect.Type, tag string) *tagInfo {
	items := strings.Split(tag, ",")
	if len(items) == 0 {
		return nil
	}
	var err error
	tagInfo := &tagInfo{}
	if limitstr, ok := findTag("limit:", items); ok {
		limitstr = strings.TrimLeftFunc(limitstr, unicode.IsSpace)
		tagInfo.Limit.Value = parseInt(limitstr[6:])
		tagInfo.Limit.Exist = true
	}

	if _, ok := findTag("require", items); ok {
		tagInfo.Require = true
	}

	if scopestr, ok := findTag("scope:", items); ok {
		scopestr = strings.TrimLeftFunc(scopestr, unicode.IsSpace)
		scopestr = strings.Trim(scopestr[6:], "[]")
		scopes := strings.Split(scopestr, " ")
		tagInfo.Scope.Items = parseScopeTag(scopes, typ)
		tagInfo.Scope.Exist = true
		tagInfo.Scope.expr = scopestr
	}

	if defstr, ok := findTag("default:", items); ok {
		defstr = strings.TrimLeftFunc(defstr, unicode.IsSpace)
		tagInfo.Default.Value, err = parseValue(typ, defstr[8:])
		tagInfo.Default.Exist = true
	}

	if patternstr, ok := findTag("pattern:", items); ok {
		patternstr = strings.TrimLeftFunc(patternstr, unicode.IsSpace)
		tagInfo.Pattern.Regexp, err = regexp.Compile(patternstr[8:])
		tagInfo.Pattern.Exist = true
	}

	if err != nil {
		panic(err)
	}

	return tagInfo
}

func parseScopeTag(scopes []string, typ reflect.Type) scopeItems {
	var items []scopeItem
	var err error
	switch typ.Kind() {
	case reflect.String:
		items = parseStringScope(scopes)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		items, err = parseIntScope(scopes)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		items, err = parseUintScope(scopes)
	case reflect.Float32, reflect.Float64:
		items, err = parseFloatScope(scopes)
	case reflect.Slice:
		sliceType := typ.Elem()
		if reflect.Struct == sliceType.Kind() {
			err = errors.New("the parameter type(" + sliceType.String() + ") is not supported")
		} else {
			items = parseScopeTag(scopes, sliceType)
		}

	default:
		err = errors.New("the type(" + typ.String() + ") parameter scope tag is not supported")
	}
	if err != nil {
		panic(err)
	}
	return items
}

func findTag(name string, items []string) (string, bool) {
	for _, v := range items {
		if strings.Contains(v, name) {
			return v, true
		}
	}
	return "", false
}

func parseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return v
}

func parseValue(typ reflect.Type, val string) (value reflect.Value, err error) {
	value = reflect.Indirect(reflect.New(typ))
	var v interface{}
	switch typ.Kind() {
	case reflect.String:
		value.SetString(val)
	case reflect.Bool:
		v, err = strconv.ParseBool(val)
		if err == nil {
			value.SetBool(v.(bool))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err = strconv.ParseInt(val, 10, 0)
		if err == nil {
			value.SetInt(v.(int64))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err = strconv.ParseUint(val, 10, 0)
		if err == nil {
			value.SetUint(v.(uint64))
		}
	case reflect.Float32:
		v, err = strconv.ParseFloat(val, 32)
		if err == nil {
			value.SetFloat(v.(float64))
		}
	case reflect.Float64:
		v, err = strconv.ParseFloat(val, 64)
		if err == nil {
			value.SetFloat(v.(float64))
		}
	default:
		err = errors.New("the parameter type is not supported")
	}
	return
}

const scopeDataTypeEqual = 1
const scopeDataTypeLessthan = 2
const scopeDataTypeGreaterthan = 3
const scopeDataTypeBetween = 4

//scope tag
//[1 2 3] or [1~100] or [0~] or [~0] or [100~] or [~100] or [~-100 -20~-10 -1 0 1 2 3 10~20 100~]
func parseScope(scope []string,
	parseData func(sv string) (interface{}, error),
	buildScopeItem func(typ int8, v ...interface{}) scopeItem) ([]scopeItem, error) {

	items := make([]scopeItem, 0, len(scope))
	for _, s := range scope {
		if len(s) == 0 {
			continue
		}
		parts := strings.SplitN(s, "~", 2)
		if len(parts) == 1 {
			v, err := parseData(parts[0])
			if err != nil {
				return nil, err
			}
			items = append(items, buildScopeItem(scopeDataTypeEqual, v))
		} else { //if len(parts) == 2 {
			if len(parts[0]) == 0 {
				v, err := parseData(parts[1])
				if err != nil {
					return nil, err
				}
				items = append(items, buildScopeItem(scopeDataTypeLessthan, v))
			} else if len(parts[1]) == 0 {
				v, err := parseData(parts[0])
				if err != nil {
					return nil, err
				}
				items = append(items, buildScopeItem(scopeDataTypeGreaterthan, v))
			} else { //len(parts[0]) > 0 && len(parts[1]) > 0
				start, err := parseData(parts[0])
				if err != nil {
					return nil, err
				}
				end, err := parseData(parts[1])
				if err != nil {
					return nil, err
				}
				items = append(items, buildScopeItem(scopeDataTypeBetween, start, end))
			}
		}
	}

	return items, nil
}

func parseIntScope(scope []string) ([]scopeItem, error) {
	return parseScope(scope, func(sv string) (interface{}, error) {
		v, err := strconv.ParseInt(sv, 10, 0)
		if err != nil {
			return nil, err
		}
		return v, nil
	}, func(typ int8, v ...interface{}) scopeItem {
		switch typ {
		case scopeDataTypeEqual:
			return intEqualScopeItem(v[0].(int64))
		case scopeDataTypeLessthan:
			return intLessthanScopeItem(v[0].(int64))
		case scopeDataTypeGreaterthan:
			return intGreaterthanScopeItem(v[0].(int64))
		case scopeDataTypeBetween:
			return intBetweenScopeItem{v[0].(int64), v[1].(int64)}
		}
		panic("type is unsupported")
	})
}

func parseUintScope(scope []string) ([]scopeItem, error) {
	return parseScope(scope, func(sv string) (interface{}, error) {
		v, err := strconv.ParseUint(sv, 10, 0)
		if err != nil {
			return nil, err
		}
		return v, nil
	}, func(typ int8, v ...interface{}) scopeItem {
		switch typ {
		case scopeDataTypeEqual:
			return uintEqualScopeItem(v[0].(uint64))
		case scopeDataTypeLessthan:
			return uintLessthanScopeItem(v[0].(uint64))
		case scopeDataTypeGreaterthan:
			return uintGreaterthanScopeItem(v[0].(uint64))
		case scopeDataTypeBetween:
			return uintBetweenScopeItem{v[0].(uint64), v[1].(uint64)}
		}
		panic("type is unsupported")
	})
}

func parseFloatScope(scope []string) ([]scopeItem, error) {
	return parseScope(scope, func(sv string) (interface{}, error) {
		v, err := strconv.ParseFloat(sv, 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	}, func(typ int8, v ...interface{}) scopeItem {
		switch typ {
		case scopeDataTypeEqual:
			return floatEqualScopeItem(v[0].(float64))
		case scopeDataTypeLessthan:
			return floatLessthanScopeItem(v[0].(float64))
		case scopeDataTypeGreaterthan:
			return floatGreaterthanScopeItem(v[0].(float64))
		case scopeDataTypeBetween:
			return floatBetweenScopeItem{v[0].(float64), v[1].(float64)}
		}
		panic("type is unsupported")
	})
}

func parseStringScope(scope []string) []scopeItem {
	return []scopeItem{stringScopeItems(scope)}
}

type scopeItem interface {
	check(v interface{}) bool
}

type scopeItems []scopeItem

func (this scopeItems) check(v interface{}) bool {
	for _, item := range []scopeItem(this) {
		if item.check(v) {
			return true
		}
	}
	return false
}

//string
type stringScopeItems []string

func (this stringScopeItems) check(v interface{}) bool {
	for _, s := range []string(this) {
		if s == v.(string) {
			return true
		}
	}
	return false
}

//int
type intEqualScopeItem int64

func (this intEqualScopeItem) check(v interface{}) bool {
	return int64(this) == toInt64(v)
}

type intBetweenScopeItem struct {
	Start int64
	End   int64
}

func (this intBetweenScopeItem) check(v interface{}) bool {
	vi64 := toInt64(v) //reflect.ValueOf(v).Int()
	return vi64 >= this.Start && vi64 <= this.End
}

type intLessthanScopeItem int64

func (this intLessthanScopeItem) check(v interface{}) bool {
	return toInt64(v) <= int64(this)
}

type intGreaterthanScopeItem int64

func (this intGreaterthanScopeItem) check(v interface{}) bool {
	return toInt64(v) >= int64(this)
}

//uint
type uintEqualScopeItem uint64

func (this uintEqualScopeItem) check(v interface{}) bool {
	return uint64(this) == toUint64(v) //reflect.ValueOf(v).Uint()
}

type uintBetweenScopeItem struct {
	Start uint64
	End   uint64
}

func (this uintBetweenScopeItem) check(v interface{}) bool {
	vu64 := toUint64(v) //reflect.ValueOf(v).Uint()
	return vu64 >= this.Start && vu64 <= this.End
}

type uintLessthanScopeItem uint64

func (this uintLessthanScopeItem) check(v interface{}) bool {
	return toUint64(v) <= uint64(this)
}

type uintGreaterthanScopeItem uint64

func (this uintGreaterthanScopeItem) check(v interface{}) bool {
	return toUint64(v) >= uint64(this)
}

//float
type floatEqualScopeItem float64

func (this floatEqualScopeItem) check(v interface{}) bool {
	return float64(this) == toFloat64(v) //reflect.ValueOf(v).Float()
}

type floatBetweenScopeItem struct {
	Start float64
	End   float64
}

func (this floatBetweenScopeItem) check(v interface{}) bool {
	vf64 := toFloat64(v) //reflect.ValueOf(v).Float()
	return vf64 >= this.Start && vf64 <= this.End
}

type floatLessthanScopeItem float64

func (this floatLessthanScopeItem) check(v interface{}) bool {
	return toFloat64(v) <= float64(this)
}

type floatGreaterthanScopeItem float64

func (this floatGreaterthanScopeItem) check(v interface{}) bool {
	return toFloat64(v) >= float64(this)
}
