package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tryor/trygo"
)

/**
 * 演示所有渲染方式
 */

func main() {
	//render text
	trygo.Get("/render/text", func(ctx *trygo.Context) {
		ctx.Render("hello world").Text()
	})

	//render html
	trygo.Get("/render/html", func(ctx *trygo.Context) {
		ctx.Render("<html><body><font size=\"6\" color=\"red\">hello world</font></body></html>").Html()
	})

	//render json
	trygo.Get("/render/json", func(ctx *trygo.Context) {
		ctx.Render([]byte("{\"id\":2,\"name\":\"John\"}")).Json()
	})

	//render jsonp
	trygo.Get("/render/jsonp", func(ctx *trygo.Context) {
		ctx.Render([]byte("{\"id\":2,\"name\":\"John\"}")).
			Jsonp(
				ctx.Input.GetValue(ctx.App.Config.Render.JsoncallbackParamName), //由前端决定是否JsonCallback格式输出数据
			)
	})

	//render xml
	trygo.Get("/render/xml", func(ctx *trygo.Context) {
		ctx.Render([]byte("<user id=\"2\" name=\"John\" />")).Xml().Nowrap()
	})

	//render template
	trygo.Get("/render/template", func(ctx *trygo.Context) {
		id := ctx.Input.GetValue("id")
		name := ctx.Input.GetValue("name")

		tplNames := "admin/index.tpl" //相对trygo.DefaultApp.SetViewsPath()设置的位置
		data := make(map[interface{}]interface{})
		data["id"] = id
		data["name"] = name
		ctx.RenderTemplate(tplNames, data)
	})

	//render gzip
	trygo.Get("/render/gzip", func(ctx *trygo.Context) {
		data := strings.Repeat("gzip demo,", 100)
		ctx.Render(data).Text().
			Gzip() //如果要默认支持Gzip，可修改配置 App.Config.Render.Gzip = true
	})

	//render struct
	trygo.Get("/render/struct", func(ctx *trygo.Context) {
		user := &User{Id: 123, Account: "demo001", Name: "demo", Sex: 1, Age: 18, Email: "demo@qq.com", Createtime: time.Now()}
		ctx.Render(user)
	})

	//render slice
	trygo.Get("/render/slice", func(ctx *trygo.Context) {

		users := make([]User, 0)
		for i := 1; i < 10; i++ {
			user := User{Id: int64(i), Account: "demo" + strconv.Itoa(i), Name: "demo", Sex: 1, Age: 18, Email: "demo@qq.com", Createtime: time.Now()}
			users = append(users, user)
		}
		ctx.Render(users)
	})

	//render page
	trygo.Get("/render/page", func(ctx *trygo.Context) {
		users := make([]User, 0)
		for i := 1; i < 10; i++ {
			user := User{Id: int64(i), Account: "demo" + strconv.Itoa(i), Name: "demo", Sex: 1, Age: 18, Email: "demo@qq.com", Createtime: time.Now()}
			users = append(users, user)
		}

		page := &page{Pno: 1, Psize: 10, Total: 100, Data: users}

		ctx.Render(page)
	})

	//render wrap success
	trygo.Get("/render/wrap/success", func(ctx *trygo.Context) {
		ctx.Render("ok").
			Wrap()
	})

	//render wrap error
	trygo.Get("/render/wrap/error", func(ctx *trygo.Context) {
		//		panic(*trygo.NewErrorResult(trygo.ERROR_CODE_PARAM_ILLEGAL, trygo.ERROR_INFO_MAP[trygo.ERROR_CODE_PARAM_ILLEGAL]))
		ctx.Render("error info").
			Wrap(trygo.ERROR_CODE_PARAM_ILLEGAL).Status(404)
	})

	//render file
	trygo.Get("/render/file", func(ctx *trygo.Context) {
		ctx.RenderFile("D:\\Go\\api\\go1.txt").Gzip()
	})

	//render stream
	trygo.Get("/render/stream", func(ctx *trygo.Context) {
		ctx.Render(strings.NewReader(strings.Repeat("stream... ", 1024))).Wrap().Text()
	})

	//set auto wrap
	trygo.Get("/render/wrap/set/(?P<auto>[^/]+)$", func(ctx *trygo.Context) {

		ctx.Input.Bind(&trygo.DefaultApp.Config.Render.Wrap, "auto")

		ctx.Render("<script>alert(\"ok\");window.location=\"/\";</script>").Html()
	})

	//set auto parse result wrap format
	trygo.Get("/render/wrap/format/autoparse/(?P<auto>[^/]+)$", func(ctx *trygo.Context) {

		ctx.Input.Bind(&trygo.DefaultApp.Config.Render.AutoParseFormat, "auto")

		ctx.Render("<script>alert(\"ok\");window.location=\"/\";</script>").Html()
	})

	//设置静态文件根位置
	trygo.SetStaticPath("/", "static/webcontent/")

	//设置模板文件根位置, 相对或绝对路径
	trygo.SetViewsPath("static/templates/")

	trygo.DefaultApp.Config.Render.AutoParseFormat = true
	//trygo.DefaultApp.Config.Render.Gzip = true
	//trygo.DefaultApp.Config.Render.Wrap = true

	fmt.Println("HTTP ListenAndServe AT ", trygo.DefaultApp.Config.Listen.Addr)
	trygo.Run()
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
