package userService

import (
	"fmt"
	"fun"
	"fun/test/Dependency"
)

type UserService struct {
	fun.Ctx
	Config *Dependency.Config
}

type User struct {
	User   string `validate:"min=6,max=10"`
	Name   *string
	Config *[]Dependency.Config
}

func (ctx UserService) HalloWord(user User) *int8 {
	fmt.Println(ctx.Config.Page, 22223)
	//panic(666)
	return nil
}

func (ctx UserService) HalloWord1(proxyClose fun.ProxyClose) *User {
	ctx.Ctx.Push(ctx.Id, ctx.RequestId, &User{
		User: "111",
	})
	proxyClose(func() {
		fmt.Println("我关闭了")
	})
	return nil
}

func (ctx UserService) HalloWord3() {
}

func init() {
	fun.BindService(UserService{})
}

type Qqq struct {
	Config *Dependency.Config
}

type Qqq1 struct {
	Config *Dependency.Config
}
