package ssss

import (
	"reflect"
	"strings"
	"testing"
)

func Test_ParseTag(t *testing.T) {
	//"name,limit:20,scope:[1 2 3],default:1,require,pattern:regex"
	var str string
	var i64 int64
	var f64 float64

	taginfos := make(Taginfos)
	taginfos.Parse(reflect.String, `field:"teststr,limit:10,scope:[One Two Three],default:Two,require"`)

	strTaginfo := taginfos.Get("teststr") //parseOneTag(reflect.String, `field:"teststr,limit:10,scope:[One Two Three],default:Two,require"`)

	if !strTaginfo.Default.Exist || strTaginfo.Default.Value.String() != "Two" {
		t.Errorf("string tag: default value does not exist,  must exist and value is Two")
	}

	if !strTaginfo.Limit.Exist || strTaginfo.Limit.Value != 10 {
		t.Errorf("string tag: limit does not exist, must exist and value is 10")
	}

	if !strTaginfo.Require {
		t.Errorf("string tag: require not is true ")
	}

	if !strTaginfo.Scope.Exist || !(strTaginfo.Scope.Items.check("One") && strTaginfo.Scope.Items.check("Two") && strTaginfo.Scope.Items.check("Three")) {
		t.Errorf("string tag: scope does not exist, must exist and value is [One Two Three]")
	}

	str = "Four"
	if err := strTaginfo.Check(str); err == nil {
		t.Errorf("string tag: %v", err)
	}

	str = ""
	if err := strTaginfo.Check(str); err == nil {
		t.Errorf("string tag: value is empty, tag is require")
	}

	str = "Three"
	var str2 string
	if err := strTaginfo.Check(str, &str2); err != nil || str2 != str {
		t.Errorf("string tag: %v, str2:%v, str:%v", err, str2, str)
	}

	taginfos.Parse(reflect.Int64, "testi64,limit:3,scope:[~-100 -2 -1 0 1 2 100~],default:1,require")
	i64Taginfo := taginfos.Get("testi64")

	if err := i64Taginfo.Check(""); err == nil || err.Error() != "tag require: value is empty, tag:testi64(int64)" {
		t.Errorf("int64 tag: %v", err)
	}

	if err := i64Taginfo.Check("5000"); err == nil || err.Error() != "tag limit: value is too long, limit:3, tag:testi64(int64)" {
		t.Errorf("int64 tag: %v", err)
	}
	if err := i64Taginfo.Check(int16(5000)); err == nil || err.Error() != "tag limit: value is too long, limit:3, tag:testi64(int64)" {
		t.Errorf("int64 tag: %v", err)
	}

	if err := i64Taginfo.Check("6"); err == nil || err.Error() != "tag scope: value is not in scope:[~-100 -2 -1 0 1 2 100~], tag:testi64(int64)" {
		t.Errorf("int64 tag: %v", err)
	}
	if err := i64Taginfo.Check(int8(6)); err == nil || err.Error() != "tag scope: value is not in scope:[~-100 -2 -1 0 1 2 100~], tag:testi64(int64)" {
		t.Errorf("int64 tag: %v", err)
	}

	if err := i64Taginfo.Check("1"); err != nil {
		t.Errorf("int64 tag: %v", err)
	}
	if err := i64Taginfo.Check(int32(1)); err != nil {
		t.Errorf("int64 tag: %v", err)
	}
	if err := i64Taginfo.Check(int64(999)); err != nil {
		t.Errorf("int64 tag: %v", err)
	}

	if err := i64Taginfo.Check("300", &i64); err != nil || i64 != 300 {
		t.Errorf("int64 tag: %v", err)
	}

	if err := i64Taginfo.Check(int64(999), &i64); err != nil || i64 != 999 {
		t.Errorf("int64 tag: %v", err)
	}

	var i16 int16
	if err := i64Taginfo.Check(250, &i16); err != nil || i16 != 250 {
		t.Errorf("int64 tag: %v", err)
	}

	taginfos.Parse(reflect.Float64, "testf64,limit:6,scope:[100.10 200.20~500.50],default:250.25")
	f64Taginfo := taginfos.Get("testf64")

	if err := f64Taginfo.Check("", &f64); err != nil || f64 != 250.25 {
		t.Errorf("float64 tag: %v, f64:%v", err, f64)
	}

	if err := f64Taginfo.Check("443.45", &f64); err != nil || f64 != 443.45 {
		t.Errorf("float64 tag: %v, f64:%v", err, f64)
	}

	if err := f64Taginfo.Check("A443.4", &f64); err == nil || err.Error() != "strconv.ParseFloat: parsing \"A443.4\": invalid syntax" {
		t.Errorf("float64 tag: %v, f64:%v", err, f64)
	}

	if err := f64Taginfo.Check(float32(343.68), &f64); err != nil || f64 != 343.68 {
		t.Errorf("float64 tag: %v, f64:%v", err, f64)
	}

	if err := f64Taginfo.Check(int(235), &f64); err != nil || f64 != 235 {
		t.Errorf("float64 tag: %v, f64:%v", err, f64)
	}

	taginfos.Parse(reflect.String, "email,limit:30,require,pattern:\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*")
	emailTaginfo := taginfos.Get("email")
	if err := emailTaginfo.Check(""); err == nil || err.Error() != "tag require: value is empty, tag:email(string)" {
		t.Errorf("email tag: %v", err)
	}

	if err := emailTaginfo.Check("trywen@qq.com", &str); err != nil || str != "trywen@qq.com" {
		t.Errorf("email tag: %v, str:%v", err, str)
	}

}

func Test_ParseStruct(t *testing.T) {
	taginfos := make(Taginfos)

	type UserForm struct {
		Account string  `field:"account,limit:20,require"`
		Pwd     string  `field:"pwd,limit:20,require"`
		Name    string  `field:"name,limit:20"`
		Sex     int     `field:"sex,scope:[1 2 3],default:1"`
		Age     int     `field:"age,scope:[0~200],default:0"`
		Email   string  `field:"email,limit:30,pattern:\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*"`
		Stature float32 `field:"stature,scope:[0.0~],default:0.0"`
	}

	type QueryUserForm struct {
		Status  []int `field:"status,scope:[1 2 3 4]"`
		Orderby []string
		Psize   int `field:"psize,scope:[10 20 50 100]"`
		Pno     int `field:"pno,scope:[1~]"`
	}

	taginfos.ParseStruct("user", reflect.TypeOf(UserForm{}), true)
	accountTaginfo := taginfos.Get("user.account")

	if accountTaginfo == nil || accountTaginfo.String() != "name:user.account(string) (limit:{{true} 20}, require:true, scope:, default:{{false} <invalid Value>}, pattern:{{false} <nil>})" {
		t.Errorf("ParseStruct error, formDomainModel=true, accountTaginfo:%v", accountTaginfo)
	}
	sexTaginfo := taginfos.Get("user.sex")
	if sexTaginfo == nil || sexTaginfo.String() != "name:user.sex(int) (limit:{{false} 0}, require:false, scope:1 2 3, default:{{true} <int Value>}, pattern:{{false} <nil>})" {
		t.Errorf("ParseStruct error, formDomainModel=true, sexTaginfo:%v", sexTaginfo)
	}

	emailTaginfo := taginfos.Get("user.email")
	if emailTaginfo == nil {
		t.Errorf("ParseStruct error, formDomainModel=true, emailTaginfo:%v", emailTaginfo)
	}
	if err := emailTaginfo.Check("trywen@qq.com"); err != nil {
		t.Errorf("ParseStruct error, err:%v", err)
	}
	if err := emailTaginfo.Check("trywen@qq"); err == nil || !strings.HasPrefix(err.Error(), "tag pattern: pattern match fail!") {
		t.Errorf("ParseStruct error, err:%v", err)
	}

	taginfos.ParseStruct("user", reflect.TypeOf(QueryUserForm{}), true)
	statusTaginfo := taginfos.Get("user.status")
	if statusTaginfo == nil || statusTaginfo.String() != "name:user.status(int) (limit:{{false} 0}, require:false, scope:1 2 3 4, default:{{false} <invalid Value>}, pattern:{{false} <nil>})" {
		t.Errorf("ParseStruct error, formDomainModel=true, statusTaginfo:%v", statusTaginfo)
	}

	var status int
	if err := statusTaginfo.Check("2", &status); err != nil || status != 2 {
		t.Errorf("ParseStruct error, err:%v", err)
	}

}
