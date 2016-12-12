package trygo

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//`field:"name,limit:20,scope:[1 2 3],default:1,require,pattern:regex"`
//key: is tag item name, value: indicates whether there is value
var supportedTagitems = map[string]bool{
	"limit":   true,
	"scope":   true,
	"default": true,
	"require": false,
	"pattern": true,
}

type Taginfos map[string]*tagInfo

func (this Taginfos) Get(name string) *tagInfo {
	return this[name]
}

func (this Taginfos) Set(name string, taginfo *tagInfo) {
	this[name] = taginfo
}

func (this Taginfos) Adds(taginfos Taginfos) {
	for k, v := range taginfos {
		this[k] = v
	}
}

func (this Taginfos) Parse(kind reflect.Kind, tag string) {
	if tagInfo := parseOneTag(kind, tag); tagInfo != nil {
		this.Set(tagInfo.Name, tagInfo)
	}
}

func (this Taginfos) ParseStruct(name string, structType reflect.Type, formDomainModel ...bool) {
	parseTag(name, structType, "", len(formDomainModel) > 0 && formDomainModel[0], this)
	//this.Adds(parseStructTag(name, structType, formDomainModel...))
}

func (this Taginfos) ParseTags(names map[string]reflect.Type, formDomainModel bool, tags ...string) {
	this.Adds(parseTags(names, tags, formDomainModel))
}

//`field:"name,limit:20,scope:[1 2 3],default:1,require,pattern:regex"`
type tagInfo struct {
	Name     string       //tag name
	TypeKind reflect.Kind //type kind value
	Limit    limitTag
	Require  bool
	Scope    scopeTag
	Default  defaultTag
	Pattern  patternTag
}

func (this *tagInfo) String() string {
	return fmt.Sprintf("name:%v(%v) (limit:%v, require:%v, scope:%v, default:%v, pattern:%v)", this.Name, this.TypeKind.String(), this.Limit, this.Require, this.Scope, this.Default, this.Pattern)
}

//v       - 可以是string类型的请求参数
//dest[0] - 如果提供此参数，将会在dest[0]中返回值，dest[0]原数据类型必须与tagInfo.Type类型一致
func (this *tagInfo) Check(v interface{}, dest ...interface{}) error {
	rval := reflect.Indirect(reflect.ValueOf(v))
	rvalType := rval.Type()

	rvalTypeKind := rvalType.Kind()
	tagTypeKind := this.TypeKind

	var valLen int
	var strval string
	if rvalType.Kind() == reflect.String {
		strval = rval.String()
		valLen = len(strval)
	} else {
		strval = fmt.Sprint(v)
		valLen = len(strval)
	}

	//Logger.Debug("v:%v", v)
	//Logger.Debug("tagInfo:%v", this.String())
	//Logger.Debug("strval:%v", strval)
	//Logger.Debug("valLen:%v", valLen)
	//Logger.Debug("strval:%v", strval)

	var destValue reflect.Value
	var destTypeKind reflect.Kind
	if len(dest) > 0 {
		if destPtr, ok := dest[0].(*reflect.Value); ok {
			destValue = *destPtr
		} else {
			destValue = reflect.ValueOf(dest[0])
			if destValue.Kind() != reflect.Ptr {
				return errors.New("dest is not pointer type")
			}
			destValue = destValue.Elem()
		}
		destTypeKind = destValue.Kind()
	}

	if valLen == 0 {
		//check require
		if this.Require {
			return errors.New("tag require: value is empty, tag:" + this.Name + "(" + this.TypeKind.String() + ")")
		} else if this.Default.Exist {
			if len(dest) > 0 {
				destValue.Set(this.Default.Value)
			}
		}
		return nil
	}

	//check length limit
	if this.Limit.Exist && this.Limit.Value > 0 && valLen > this.Limit.Value {
		return errors.New(fmt.Sprintf("tag limit: value is too long, limit:%v, tag:%v(%v)", this.Limit.Value, this.Name, this.TypeKind.String()))
	}

	//check pattern
	if this.Pattern.Exist {
		if !this.Pattern.Regexp.MatchString(strval) {
			return errors.New(fmt.Sprintf("tag pattern: pattern match fail!, pattern:%v, tag:%v(%v)", this.Pattern.Regexp.String(), this.Name, this.TypeKind.String()))
		}
	}

	//	Logger.Debug("rvalTypeKind:%v", rvalTypeKind)
	//	Logger.Debug("destTypeKind:%v", destTypeKind)
	//	Logger.Debug("tagTypeKind:%v", tagTypeKind)

	var val interface{}
	if len(dest) > 0 {
		if rvalTypeKind != destTypeKind {
			var err error
			_, err = parseValue(strval, destTypeKind, &destValue)
			if err != nil {
				return err
			}
			val = destValue.Interface() //rval.Interface()
		} else {
			destValue.Set(rval)
			val = v
		}
	}

	if val == nil {
		if rvalTypeKind != tagTypeKind {
			value, err := parseValue(strval, tagTypeKind, nil)
			if err != nil {
				return err
			}
			val = value.Interface()
		} else {
			val = v
		}
	}

	if this.Scope.Exist && !this.Scope.Items.check(val) {
		return errors.New(fmt.Sprintf("tag scope: value is not in scope:[%v], tag:%v(%v)", this.Scope.String(), this.Name, this.TypeKind.String()))
	}

	return nil
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
func parseTags(pnametypes map[string]reflect.Type, tags []string, formDomainModel bool) (formatedTags Taginfos) {
	if len(pnametypes) == 0 {
		return
	}

	pnames := make(map[string]reflect.Type)
	for k, v := range pnametypes {
		pnames[k] = v
	}

	formatedTags = make(Taginfos)
	for _, tag := range tags {
		name, tag := pretreatTag(tag)
		if typ, ok := pnames[name]; ok {
			parseTag(name, typ, tag, formDomainModel, formatedTags)
			delete(pnames, name)
		}
	}

	for name, typ := range pnames {
		parseTag(name, typ, "", formDomainModel, formatedTags)
	}
	return
}

func pretreatTag(tag string) (name, rettag string) {
	pair := strings.SplitN(tag, ":\"", 2)
	if len(pair) > 1 {
		tag = pair[1]
		if tag[len(tag)-1] == '"' {
			tag = tag[0 : len(tag)-1]
		}
	}
	pair = strings.SplitN(tag, ",", 2)
	return strings.TrimSpace(pair[0]), strings.TrimSpace(tag)
}

//分析tag信息
//typ  - 属性类型
//tag - 属性tag信息, `:"name,limit:10,scope:[One Two Three],default:Two,require"`
func parseOneTag(kind reflect.Kind, tag string) *tagInfo {
	stag := tag
	_, tag = pretreatTag(tag)
	if tag == "" {
		panic("tag is empty, tag:" + stag)
	}
	//kind := typ.Kind()
	switch kind {
	case reflect.Slice:
		panic("ssss: the slice type is not supported, @see ParseStructTag()")
	case reflect.Struct:
		panic("ssss: the struct type is not supported, @see ParseStructTag()")
	}

	if taginfo := parseTagItems(kind, tag); taginfo != nil {
		return taginfo
	} else {
		panic("parse fail, tag:" + stag)
	}

	return nil
}

//分析tag信息，如果是结构类型，将自动分析结构tag的field字段属性
//name - 属性名
//typ  - 属性类型
//tag - 属性tag信息
//formDomainModel[0] @see config.FormDomainModel
//@return taginfos - 在taginfos中返回格式化后的tag信息

//func parseStructTag(name string, structType reflect.Type, formDomainModel ...bool) (taginfos Taginfos) {
//	taginfos = make(Taginfos)
//	parseTag(name, structType, "", len(formDomainModel) > 0 && formDomainModel[0], taginfos)
//	return
//}

func parseTag(pname string, ptype reflect.Type, tag string, formDomainModel bool, formatedTags Taginfos) {
	kind := ptype.Kind()
	if kind == reflect.Struct {
		for i := 0; i < ptype.NumField(); i++ {
			stag := tag
			f := ptype.Field(i)
			var attrName string
			if f.Anonymous {
				attrName = pname
			} else {
				paramTag := strings.TrimSpace(f.Tag.Get("field"))
				if paramTag == "-" {
					continue
				}
				if paramTag != "" {
					stag = paramTag
				}
				paramTags := strings.SplitN(paramTag, ",", 2)
				if len(paramTag) > 0 {
					attrName = strings.TrimSpace(paramTags[0])
					if attrName == "-" {
						continue
					}
				}
				if len(attrName) == 0 {
					attrName = strings.ToLower(f.Name[0:1]) + f.Name[1:]
				}
				if formDomainModel && pname != "" {
					attrName = pname + "." + attrName
				}
			}
			parseTag(attrName, f.Type, stag, formDomainModel, formatedTags)
		}

	} else if kind == reflect.Slice {
		parseTag(pname, ptype.Elem(), tag, formDomainModel, formatedTags)
	} else {
		if len(tag) > 0 {
			if taginfo := parseTagItems(kind, tag); taginfo != nil {
				taginfo.Name = pname
				formatedTags[pname] = taginfo
			}
		}
	}

}

func parseTotagitemmap(tag string) (string, map[string]string) {
	items := strings.Split(tag, ",")
	if len(items) == 0 {
		panic("tag is empty or format error, tag:" + tag)
	}
	tagitemmap := make(map[string]string)
	for _, item := range items[1:] {
		pair := strings.SplitN(item, ":", 2)
		name := strings.TrimSpace(pair[0])

		hasval, ok := supportedTagitems[name]
		if !ok {
			panic("tag item is not supported, item:" + name + ", tag:" + tag)
		}
		if hasval && (len(pair) < 2 || pair[1] == "") {
			panic("tag item (" + name + ") must have value, tag:" + tag)
		}

		if len(pair) > 1 {
			tagitemmap[name] = pair[1]
		} else {
			tagitemmap[name] = ""
		}
	}
	return strings.TrimSpace(items[0]), tagitemmap
}

func parseTagItems(kind reflect.Kind, tag string) *tagInfo {

	name, tagitemmap := parseTotagitemmap(tag)
	if name == "" {
		panic("tag name is empty, tag:" + tag)
	}

	tagInfo := &tagInfo{Name: name}
	if limit, ok := tagitemmap["limit"]; ok {
		tagInfo.Limit.Value = parseInt(limit)
		tagInfo.Limit.Exist = true
	}

	if _, ok := tagitemmap["require"]; ok {
		tagInfo.Require = true
	}

	if scope, ok := tagitemmap["scope"]; ok {
		scope = strings.TrimSpace(scope)
		scope = strings.Trim(scope, "[]")
		tagInfo.Scope.Items = parseScopeTag(strings.Split(scope, " "), kind)
		tagInfo.Scope.Exist = true
		tagInfo.Scope.expr = scope
	}

	if pattern, ok := tagitemmap["pattern"]; ok {
		var err error
		tagInfo.Pattern.Regexp, err = regexp.Compile(pattern)
		if err != nil {
			panic(err)
		}
		tagInfo.Pattern.Exist = true
	}

	if def, ok := tagitemmap["default"]; ok {
		defvalue, err := parseValue(def, kind, nil) //tagInfo.Default.Value, err = parseValue(def, kind)
		if err != nil {
			panic(err)
		}
		tagInfo.Default.Value = *defvalue
		tagInfo.Default.Exist = true

		if tagInfo.Scope.Exist && !tagInfo.Scope.Items.check(tagInfo.Default.Value.Interface()) {
			panic(fmt.Sprintf("default:%v is not in scope:%s", tagInfo.Default.Value, tagInfo.Scope.expr))
		}

		if tagInfo.Pattern.Exist && !tagInfo.Pattern.Regexp.MatchString(def) {
			panic(fmt.Sprintf("default:%s cannot match pattern:%s", def, tagInfo.Pattern.Regexp.String()))
		}

	}
	tagInfo.TypeKind = kind

	return tagInfo
}

//func parseTagItems_old(typ reflect.Type, tag string) *tagInfo {

//	items := strings.Split(tag, ",")
//	if len(items) == 0 {
//		panic("tag is empty or format error, tag:" + tag)
//	}

//	var err error
//	tagInfo := &tagInfo{}
//	if limitstr, ok := findTag("limit:", items); ok {
//		limitstr = strings.TrimLeftFunc(limitstr, unicode.IsSpace)
//		tagInfo.Limit.Value = parseInt(limitstr[6:])
//		tagInfo.Limit.Exist = true
//	}

//	if _, ok := findTag("require", items); ok {
//		tagInfo.Require = true
//	}

//	if scopestr, ok := findTag("scope:", items); ok {
//		scopestr = strings.TrimLeftFunc(scopestr, unicode.IsSpace)
//		scopestr = strings.Trim(scopestr[6:], "[]")
//		scopes := strings.Split(scopestr, " ")
//		tagInfo.Scope.Items = parseScopeTag(scopes, typ)
//		tagInfo.Scope.Exist = true
//		tagInfo.Scope.expr = scopestr
//	}

//	if patternstr, ok := findTag("pattern:", items); ok {
//		patternstr = strings.TrimLeftFunc(patternstr, unicode.IsSpace)
//		tagInfo.Pattern.Regexp, err = regexp.Compile(patternstr[8:])
//		tagInfo.Pattern.Exist = true
//	}

//	if defstr, ok := findTag("default:", items); ok {
//		defstr = strings.TrimLeftFunc(defstr, unicode.IsSpace)
//		defstr = defstr[8:]
//		tagInfo.Default.Value, err = parseValue(defstr, typ)
//		tagInfo.Default.Exist = true

//		if tagInfo.Scope.Exist && !tagInfo.Scope.Items.check(tagInfo.Default.Value.Interface()) {
//			panic(fmt.Sprintf("default:%v is not in scope:%s", tagInfo.Default.Value, tagInfo.Scope.expr))
//		}

//		if tagInfo.Pattern.Exist && !tagInfo.Pattern.Regexp.MatchString(defstr) {
//			panic(fmt.Sprintf("default:%s cannot match pattern:%s", defstr, tagInfo.Pattern.Regexp.String()))
//		}

//	}

//	if err != nil {
//		panic(err)
//	}

//	tagInfo.Type = typ

//	return tagInfo
//}

func parseScopeTag(scopes []string, kind reflect.Kind) scopeItems {
	var items []scopeItem
	var err error
	switch kind {
	case reflect.String:
		items = parseStringScope(scopes)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		items, err = parseIntScope(scopes)
		if err != nil {
			panic(err)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		items, err = parseUintScope(scopes)
		if err != nil {
			panic(err)
		}
	case reflect.Float32, reflect.Float64:
		items, err = parseFloatScope(scopes)
		if err != nil {
			panic(err)
		}
		//	case reflect.Slice:
		//		sliceType := typ.Elem()
		//		if reflect.Struct == sliceType.Kind() {
		//			err = errors.New("the parameter type(" + sliceType.String() + ") is not supported")
		//		} else {
		//			items = parseScopeTag(scopes, sliceType)
		//		}

	default:
		panic("the type(" + kind.String() + ") parameter scope tag is not supported")
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
		v, err := strconv.ParseInt(sv, 10, 64)
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
		v, err := strconv.ParseUint(sv, 10, 64)
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
