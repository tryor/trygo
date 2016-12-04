package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	//	"github.com/trygo/ssss"
	"trygo/ssss"
)

/**
 * 演示所有路由方式
 */

func main() {
	go func() {
		//http://localhost:6060/debug/pprof/
		ssss.Logger.Info("%v", http.ListenAndServe("localhost:6060", nil))
	}()

	//ssss.DefaultApp.Config.RunMode = ssss.DEV

	//router, get
	ssss.Get("/router/get", func(ctx *ssss.Context) {
		p1 := ctx.Input.GetValue("p1")
		p2 := ctx.Input.GetValue("p2")
		ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
	})

	//router, post
	ssss.Post("/router/post", func(ctx *ssss.Context) {
		p1 := ctx.Input.GetValue("p1")
		p2 := ctx.Input.GetValue("p2")
		ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
	})

	//router, func
	ssss.RegisterFunc("post|get|put", "/router/func", func(ctx *ssss.Context) {
		p1 := ctx.Input.GetValue("p1")
		p2 := ctx.Input.GetValue("p2")
		ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
	})

	//router, http.Handler
	ssss.RegisterHandler("/router/handler", &Handler{})

	//router, RESTful
	ssss.RegisterRESTful("/router/restful", &RESTfulController{})

	//router, ssss normal
	ssss.Register("post", "/router/user/create", &UserController{}, "Create")
	ssss.Register("get", "/router/user/get", &UserController{}, "GetUser")

	//router, regexp pattern
	ssss.Register("*", "/router/user/query/(?P<pno>[^/]+)?/(?P<psize>[^/]+)?/(?P<orderby>[^/]+)?$", &UserController{}, "Query")

	//router, bind parameters
	ssss.Register("post", "/router/user/login", &UserController{}, "Login(account, pwd string, devid int)")

	//router, bind parameters and set qualifier tag
	ssss.Register("post", "/router/user/login2", &UserController{}, "Login(account, pwd string, devid int)", "account,limit:20,require", "pwd,limit:20,require", "devid,scope:[1 2 3 4],default:1")

	//router, bind struct parameters and set qualifier tag in struct
	ssss.Register("post", "/router/user/edit", &UserController{}, "Edit(user User)")

	//设置静态文件根位置
	ssss.SetStaticPath("/", "static/webcontent/")

	fmt.Println("HTTP ListenAndServe AT ", ssss.DefaultApp.Config.HttpPort)
	ssss.Run()

}

type Handler struct{}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := ssss.NewContext(rw, r, ssss.DefaultApp)
	p1 := ctx.Input.GetValue("p1")
	p2 := ctx.Input.GetValue("p2")
	ctx.Render("(" + ctx.Request.Method + ")hello world, p1=" + p1 + ", p2=" + p2)
}

type RESTfulController struct {
	ssss.Controller
}

func (c *RESTfulController) Get() {
	id := c.Ctx.Input.GetValue("id")
	c.Render("(" + c.Ctx.Request.Method + ")hello world, id=" + id)
}

type UserController struct {
	ssss.Controller
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
	Id         int64     `field:"-"                           json:"id,omitempty" xml:"id,attr,omitempty"`
	Account    string    `field:"account,limit:20,require"    json:"account,omitempty" xml:"account,attr,omitempty"`
	Name       string    `field:"name,limit:20,require"       json:"name,omitempty" xml:"name,attr,omitempty"`
	Pwd        string    `field:"pwd,limit:20,require"        json:"-" xml:"-"`
	Sex        int       `field:"sex,scope:[1 2 3],default:1" json:"sex,omitempty" xml:"sex,attr,omitempty"`
	Age        int       `field:"age,scope:[0~200],default:0" json:"age,omitempty" xml:"age,attr,omitempty"`
	Email      string    `field:"email,limit:30,pattern:\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*" json:"email,omitempty" xml:"email,attr,omitempty"`
	Createtime time.Time `field:"-"                           json:"createtime,omitempty" xml:"createtime,attr,omitempty"`
}

type page struct {
	Pno   int         `json:"pno" xml:"pno,attr"`     //页号
	Psize int         `json:"psize" xml:"psize,attr"` //一页数据行数
	Total int         `json:"total" xml:"total,attr"` //总数
	Data  interface{} `json:"data" xml:"data"`        //具体数据
}
