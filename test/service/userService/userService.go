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
	User   string
	Name   string
	Config Dependency.Config
}

func (ctx UserService) HalloWord() *int8 {
	fmt.Println(ctx.Config.Page, 22223)
	//panic(666)
	return nil
}

func (ctx UserService) HalloWord1(porxy fun.ProxyClose) *User {
	return nil
}

func (ctx UserService) HalloWord3() {
}

func init() {
	fun.BindService(UserService{}, Qqq{}, Qqq1{})
}

type Qqq struct {
	Config *Dependency.Config
}

type Qqq1 struct {
	Config *Dependency.Config
}

func (q Qqq) Guard(serviceName string, methodName string, state map[string]string) *fun.Result[any] {
	//TODO implement me
	fmt.Println("前面1")
	return nil
}

func (q Qqq1) Guard(serviceName string, methodName string, state map[string]string) *fun.Result[any] {
	//TODO implement me
	fmt.Println(serviceName, methodName)
	fmt.Println("前面2")
	//a := fun.Error(300, "22222")
	return nil
}
