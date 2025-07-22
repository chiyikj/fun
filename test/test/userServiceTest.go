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

func (ctx UserServiceTest) HalloWord() {
	request := fun.GetRequestInfo(map[string]any{
		"User": nil,
		"Name": "1212",
	}, map[string]string{})
	reu := fun.MockRequest[*int8](fun.GetClientInfo("123456"), request)
	fmt.Println(reu)
}

func main() {
	fmt.Println(reflect.TypeOf(userService.UserService{}).PkgPath())
	fun.Test(UserServiceTest{})
}
