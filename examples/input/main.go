package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/tryor/trygo"
)

/**
 * 演示请求参数处理
 */

func main() {
	//input, query parameters
	trygo.Get("/input/get", func(ctx *trygo.Context) {
		q1 := ctx.Input.GetValue("q1")
		q2 := ctx.Input.GetValue("q2")
		q3 := ctx.Input.GetValues("q3")
		ctx.Render(fmt.Sprintf("(%s)hello world, q1=%s, q2=%s, q3=%v", ctx.Request.Method, q1, q2, q3))
	})

	//input, query and form parameters
	trygo.Post("/input/post", func(ctx *trygo.Context) {
		q1 := ctx.Input.GetValue("q1")
		q2 := ctx.Input.GetValue("q2")
		q3 := ctx.Input.GetValues("q3")
		p1 := ctx.Input.GetValue("p1")
		p2 := ctx.Input.GetValues("p2")
		ctx.Render(fmt.Sprintf("(%s)hello world, q1=%s, q2=%s, q3=%v, p1=%s, p2=%v", ctx.Request.Method, q1, q2, q3, p1, p2))
	})

	//input, path parameters
	trygo.Get("/input/path/(?P<id>[^/]+)?$", func(ctx *trygo.Context) {
		id := ctx.Input.GetValue("id")
		ctx.Render(fmt.Sprintf("(%s)hello world, id=%s", ctx.Request.Method, id))
	})

	//input, path parameters
	trygo.Get("/input/path/(?P<year>[^/]+)?/(?P<month>[^/]+)?/(?P<day>[^/]+)?$", func(ctx *trygo.Context) {
		year := ctx.Input.GetValue("year")
		month := ctx.Input.GetValue("month")
		day := ctx.Input.GetValue("day")
		ctx.Render(fmt.Sprintf("(%s)hello world, year=%s, month=%s, day=%s", ctx.Request.Method, year, month, day))
	})

	//input, bind parameters
	trygo.Get("/input/bind", func(ctx *trygo.Context) {
		var q1 string
		var q2 int
		var q3 []float32
		ctx.Input.Bind(&q1, "q1")
		ctx.Input.Bind(&q2, "q2")
		ctx.Input.Bind(&q3, "q3")
		ctx.Render(fmt.Sprintf("(%s)hello world, q1=%s, q2=%d, q3=%v", ctx.Request.Method, q1, q2, q3))
	})

	//input, bind parameters and check format
	trygo.Get("/input/bind/checkformat", func(ctx *trygo.Context) {
		var q1 string
		var q2 int
		var q3 []float32

		taginfos := make(trygo.Taginfos)
		taginfos.Parse(reflect.String, "q1,limit:5,require")
		taginfos.Parse(reflect.Int, "q2,scope:[1 2 3],default:1")
		taginfos.Parse(reflect.Float32, "q3,scope:[-100.0~200.0],default:0")

		ctx.Input.Bind(&q1, "q1", taginfos)
		ctx.Input.Bind(&q2, "q2", taginfos)
		ctx.Input.Bind(&q3, "q3", taginfos)
		ctx.Render(fmt.Sprintf("(%s)hello world, q1=%s, q2=%d, q3=%v", ctx.Request.Method, q1, q2, q3))
	})

	//input, bind struct parameters and check format, see struct tag "param"
	trygo.Post("/input/bind/struct", func(ctx *trygo.Context) {
		user := &User{}
		ctx.Input.Bind(user, "user")
		ctx.Render(user).Json()
	})

	//input, auto bind parameters
	trygo.Register("post", "/input/bind/auto", &UserController{}, "Login(account, pwd string, devid int)")

	//input, auto bind parameters and check format
	trygo.Register("post", "/input/bind/auto/checkformat", &UserController{}, "Login(account, pwd string, devid int)", "account,limit:20,require", "pwd,limit:20,require", "devid,scope:[1 2 3 4],default:1")

	//input, auto bind struct parameters and check format
	trygo.Register("post", "/input/bind/auto/struct", &UserController{}, "Edit(user User)")

	//设置静态文件根位置
	trygo.SetStaticPath("/", "static/webcontent/")

	fmt.Println("HTTP ListenAndServe AT ", trygo.DefaultApp.Config.Listen.Addr)
	trygo.Run()

}

type UserController struct {
	trygo.Controller
}

func (c *UserController) Login(account, pwd string, devid int) {
	c.Render(fmt.Sprintf("account:%v, pwd:%v, devid:%v", account, pwd, devid))
}

func (c *UserController) Edit(user User) {
	c.Render(user).Json()
}

func (c *UserController) GetUser() {
	var id int64
	c.Ctx.Input.Bind(&id, "id")
	user := &User{Id: id, Account: "demo001"}
	c.Render(user).Json()
}

type User struct {
	Id         int64     `param:"-"                           json:"id,omitempty" xml:"id,attr,omitempty"`
	Account    string    `param:"account,limit:20,require"    json:"account,omitempty" xml:"account,attr,omitempty"`
	Name       string    `param:"name,limit:20,require"       json:"name,omitempty" xml:"name,attr,omitempty"`
	Pwd        string    `param:"pwd,limit:20,require"        json:"-" xml:"-"`
	Sex        int       `param:"sex,scope:[1 2 3],default:1" json:"sex,omitempty" xml:"sex,attr,omitempty"`
	Age        int       `param:"age,scope:[0~200],default:0" json:"age,omitempty" xml:"age,attr,omitempty"`
	Email      string    `param:"email,limit:30,pattern:\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*" json:"email,omitempty" xml:"email,attr,omitempty"`
	Createtime time.Time `param:"-"                           json:"createtime,omitempty" xml:"createtime,attr,omitempty"`
}
