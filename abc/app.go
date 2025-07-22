package main

import (
	"fmt"
	"reflect"
)

// 定义一个接口
type a1 interface {
	Name() string
}

// 定义一个结构体
type User struct {
	Name1 string
}

// 如果你想让 User 实现 a1 接口，就加上这个方法
func (u User) Name() string {
	return u.Name1
}

func main() {
	// 创建实例
	u := User{}

	// 使用反射判断是否实现了接口
	t := reflect.TypeOf(u)
	iface := reflect.TypeOf((*a1)(nil)).Elem()

	// 判断类型是否实现了接口
	if t.Implements(iface) {
		fmt.Println("User 实现了 a1 接口")
	} else {
		fmt.Println("User 没有实现 a1 接口")
	}
}
