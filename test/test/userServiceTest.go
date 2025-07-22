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
		"User": "1212",
		"Name": "22333",
	}, map[string]string{})
	result := fun.MockRequest[string](fun.GetClientInfo("123456"), request)
	fmt.Println(result)
}

func main() {
	fmt.Println(reflect.TypeOf(userService.UserService{}).PkgPath())
	fun.Test(UserServiceTest{})
}
