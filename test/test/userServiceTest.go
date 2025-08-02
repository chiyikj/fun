package main

import (
	"fmt"
	"fun"
	"fun/test/service/userService"
	"reflect"
)

type UserServiceTest struct {
	Service userService.UserService
}

//func (ctx UserServiceTest) HalloWord() {
//	request := fun.GetRequestInfo(userService.UserService{}, "HalloWord", userService.User{
//		Name: nil,
//		User: "123456",
//	}, map[string]string{})
//	fmt.Println(fun.MockRequest[*int8](request))
//}

func (ctx UserServiceTest) HalloWord1() {
	request := fun.GetRequestInfo(userService.UserService{}, "HalloWord1", map[string]string{}, map[string]string{})
	proxy := fun.ProxyMessage{
		Message: func(message any) {
			fmt.Printf("Received message: %+v\n", message)

		},
		Close: func() {
			// 处理连接关闭事件
			fmt.Println("Connection closed")
		},
	}
	fun.MockProxy(request, proxy, 10)
}

func main() {
	fmt.Println(reflect.TypeOf(userService.UserService{}).PkgPath())
	fun.Test(UserServiceTest{})
}
