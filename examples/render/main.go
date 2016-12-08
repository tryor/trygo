package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/trygo/ssss"
)

/**
 * 演示所有渲染方式
 */

func main() {
	//render text
	ssss.Get("/render/text", func(ctx *ssss.Context) {
		ctx.Render("hello world").Text()
	})

	//render html
	ssss.Get("/render/html", func(ctx *ssss.Context) {
		ctx.Render("<html><body><font size=\"6\" color=\"red\">hello world</font></body></html>").Html()
	})

	//render json
	ssss.Get("/render/json", func(ctx *ssss.Context) {
		ctx.Render([]byte("{\"id\":2,\"name\":\"John\"}")).Json()
	})

	//render jsoncallback
	ssss.Get("/render/jsoncallback", func(ctx *ssss.Context) {
		ctx.Render([]byte("{\"id\":2,\"name\":\"John\"}")).
			JsonCallback(
				ctx.Input.GetValue(ctx.App.Config.Render.JsoncallbackParamName), //由前端决定是否JsonCallback格式输出数据
			)
	})

	//render xml
	ssss.Get("/render/xml", func(ctx *ssss.Context) {
		ctx.Render([]byte("<user id=\"2\" name=\"John\" />")).Xml()
	})

	//render template
	ssss.Get("/render/template", func(ctx *ssss.Context) {
		id := ctx.Input.GetValue("id")
		name := ctx.Input.GetValue("name")

		tplNames := "admin/index.tpl" //相对ssss.DefaultApp.SetViewsPath()设置的位置
		data := make(map[interface{}]interface{})
		data["id"] = id
		data["name"] = name
		ctx.RenderTemplate(tplNames, data)
	})

	//render gzip
	ssss.Get("/render/gzip", func(ctx *ssss.Context) {
		data := strings.Repeat("gzip demo,", 100)
		ctx.Render(data).Text().
			Gzip() //如果要默认支持Gzip，可修改配置 App.Config.Render.Gzip = true
	})

	//render struct
	ssss.Get("/render/struct", func(ctx *ssss.Context) {
		user := &User{Id: 123, Account: "demo001", Name: "demo", Sex: 1, Age: 18, Email: "demo@qq.com", Createtime: time.Now()}
		ctx.Render(user)
	})

	//render slice
	ssss.Get("/render/slice", func(ctx *ssss.Context) {

		users := make([]User, 0)
		for i := 1; i < 10; i++ {
			user := User{Id: int64(i), Account: "demo" + strconv.Itoa(i), Name: "demo", Sex: 1, Age: 18, Email: "demo@qq.com", Createtime: time.Now()}
			users = append(users, user)
		}
		ctx.Render(users)
	})

	//render page
	ssss.Get("/render/page", func(ctx *ssss.Context) {
		users := make([]User, 0)
		for i := 1; i < 10; i++ {
			user := User{Id: int64(i), Account: "demo" + strconv.Itoa(i), Name: "demo", Sex: 1, Age: 18, Email: "demo@qq.com", Createtime: time.Now()}
			users = append(users, user)
		}

		page := &page{Pno: 1, Psize: 10, Total: 100, Data: users}

		ctx.Render(page)
	})

	//render wrap success
	ssss.Get("/render/wrap/success", func(ctx *ssss.Context) {
		ctx.Render("ok").
			Wrap()
	})

	//render wrap error
	ssss.Get("/render/wrap/error", func(ctx *ssss.Context) {
		//		panic(*ssss.NewErrorResult(ssss.ERROR_CODE_PARAM_ILLEGAL, ssss.ERROR_INFO_MAP[ssss.ERROR_CODE_PARAM_ILLEGAL]))
		ctx.Render("error info").
			Wrap(ssss.ERROR_CODE_PARAM_ILLEGAL).Status(404)
	})

	//render file
	ssss.Get("/render/file", func(ctx *ssss.Context) {
		ctx.RenderFile("D:\\Go\\api\\go1.txt").Gzip()
	})

	//set auto wrap
	ssss.Get("/render/wrap/set/(?P<auto>[^/]+)$", func(ctx *ssss.Context) {

		ctx.Input.Bind(&ssss.DefaultApp.Config.Render.Wrap, "auto")

		ctx.Render("<script>alert(\"ok\");window.location=\"/\";</script>").Html()
	})

	//set auto parse result wrap format
	ssss.Get("/render/wrap/format/autoparse/(?P<auto>[^/]+)$", func(ctx *ssss.Context) {

		ctx.Input.Bind(&ssss.DefaultApp.Config.Render.AutoParseFormat, "auto")

		ctx.Render("<script>alert(\"ok\");window.location=\"/\";</script>").Html()
	})

	//设置静态文件根位置
	ssss.SetStaticPath("/", "static/webcontent/")

	//设置模板文件根位置, 相对或绝对路径
	ssss.SetViewsPath("static/templates/")

	ssss.DefaultApp.Config.Render.AutoParseFormat = true
	//ssss.DefaultApp.Config.Render.Gzip = true
	//ssss.DefaultApp.Config.Render.Wrap = true

	fmt.Println("HTTP ListenAndServe AT ", ssss.DefaultApp.Config.Listen.Addr)
	ssss.Run()
}

type User struct {
	Id         int64     `json:"id,omitempty" xml:"id,attr,omitempty"`
	Account    string    `json:"account,omitempty" xml:"account,attr,omitempty"`
	Name       string    `json:"name,omitempty" xml:"name,attr,omitempty"`
	Pwd        string    `json:"-" xml:"-"`
	Sex        int       `json:"sex,omitempty" xml:"sex,attr,omitempty"`
	Age        int       `json:"age,omitempty" xml:"age,attr,omitempty"`
	Email      string    `json:"email,omitempty" xml:"email,attr,omitempty"`
	Createtime time.Time `json:"createtime,omitempty" xml:"createtime,attr,omitempty"`
}

func (u *User) String() string {
	return fmt.Sprintf("{%d,%s,%s,%d,%d,%s,%v}", u.Id, u.Account, u.Name, u.Sex, u.Age, u.Email, u.Createtime)
}

type page struct {
	Pno   int         `json:"pno" xml:"pno,attr"`
	Psize int         `json:"psize" xml:"psize,attr"`
	Total int         `json:"total" xml:"total,attr"`
	Data  interface{} `json:"data" xml:"data"`
}
