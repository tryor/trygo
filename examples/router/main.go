package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/tryor/trygo"
)

/**
 * 演示所有路由方式
 */

func main() {
	//router, get
	trygo.Get("/router/get", func(ctx *trygo.Context) {
		p1 := ctx.Input.GetValue("p1")
		p2 := ctx.Input.GetValue("p2")
		ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
	})

	//router, post
	trygo.Post("/router/post", func(ctx *trygo.Context) {
		p1 := ctx.Input.GetValue("p1")
		p2 := ctx.Input.GetValue("p2")
		ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
	})

	//router, func
	trygo.RegisterFunc("post|get|put", "/router/func", func(ctx *trygo.Context) {
		p1 := ctx.Input.GetValue("p1")
		p2 := ctx.Input.GetValue("p2")
		ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
	})

	//router, http.Handler
	trygo.RegisterHandler("/router/handler", &Handler{})

	//router, RESTful
	trygo.RegisterRESTful("/router/restful", &RESTfulController{})

	//router, trygo normal
	trygo.Register("post", "/router/user/create", &UserController{}, "Create")
	trygo.Register("get", "/router/user/get", &UserController{}, "GetUser")

	//router, regexp pattern
	trygo.Register("*", "/router/user/query/(?P<pno>[^/]+)?/(?P<psize>[^/]+)?/(?P<orderby>[^/]+)?$", &UserController{}, "Query")

	//router, bind parameters
	trygo.Register("post", "/router/user/login", &UserController{}, "Login(account, pwd string, devid int)")

	//router, bind parameters and set qualifier tag
	trygo.Register("post", "/router/user/login2", &UserController{}, "Login(account, pwd string, devid int)", "account,limit:20,require", "pwd,limit:20,require", "devid,scope:[1 2 3 4],default:1")

	//router, bind struct parameters and set qualifier tag in struct
	trygo.Register("post", "/router/user/edit", &UserController{}, "Edit(user User)")

	//设置静态文件根位置
	trygo.SetStaticPath("/", "static/webcontent/")

	fmt.Println("HTTP ListenAndServe AT ", trygo.DefaultApp.Config.Listen.Addr)
	trygo.Run()

}

type Handler struct{}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := trygo.NewContext(rw, r, trygo.DefaultApp)
	p1 := ctx.Input.GetValue("p1")
	p2 := ctx.Input.GetValue("p2")
	ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
}

type RESTfulController struct {
	trygo.Controller
}

func (c *RESTfulController) Get() {
	id := c.Ctx.Input.GetValue("id")
	c.Render("(" + c.Ctx.Request.Method + ")hello world, id=" + id)
}

type UserController struct {
	trygo.Controller
}

func (c *UserController) Login(account, pwd string, devid int) {
	c.Render(fmt.Sprintf("account:%v, pwd:%v, devid:%v", account, pwd, devid))
}

func (c *UserController) Create() {
	user := &User{}
	c.Ctx.Input.Bind(user, "user")
	c.Render(user).Json()
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

func (c *UserController) Query() {
	var pno, psize int
	var orderby string
	c.Ctx.Input.Bind(&pno, "pno")
	c.Ctx.Input.Bind(&psize, "psize")
	c.Ctx.Input.Bind(&orderby, "orderby")

	users := make([]User, 0)
	for i := 1; i < psize; i++ {
		user := User{Id: int64(i), Account: "demo" + strconv.Itoa(i), Name: "demo", Sex: 1, Age: 18, Email: "demo@qq.com"}
		users = append(users, user)
	}

	page := &page{Pno: pno, Psize: psize, Total: 100, Data: users}

	c.Render(page).Json()
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

type page struct {
	Pno   int         `json:"pno" xml:"pno,attr"`
	Psize int         `json:"psize" xml:"psize,attr"`
	Total int         `json:"total" xml:"total,attr"`
	Data  interface{} `json:"data" xml:"data"`
}
