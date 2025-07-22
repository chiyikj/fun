package fun

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
)

func Test(TestEntity any) {
	t := reflect.TypeOf(TestEntity)
	// 检查是否为结构体
	if t.Kind() != reflect.Struct {
		panic("输入必须是结构体类型")
	}

	// 检查字段数量是否为1
	if t.NumField() != 1 {
		panic(fmt.Sprintf("结构体必须只有一个字段，当前有 %d 个字段", t.NumField()))
	}

	// 获取唯一的字段
	field := t.Field(0)

	// 检查字段名是否为Service
	if field.Name != "Service" {
		panic(fmt.Sprintf("字段名必须为Service，当前为 %s", field.Name))
	}

	// 检查是否为匿名字段
	if field.Anonymous {
		panic("Service字段不能是匿名字段")
	}
	serviceType := t.Field(0).Type
	total := serviceType.NumMethod()
	var methodList []string
	for i := 0; i < total; i++ {
		method := serviceType.Method(i)
		methodList = append(methodList, method.Name)
	}
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		method.Func.Call([]reflect.Value{reflect.ValueOf(TestEntity)})
		methodList = slices.DeleteFunc(methodList, func(s string) bool {
			return s == method.Name // 删除奇数
		})

	}
	num := float64(100) / float64(total)
	coverageRate := uint8(float64(total-len(methodList)) * num)
	testInfo := TestInfo{
		Service:      serviceType.Name(),
		CoverageRate: strconv.Itoa(int(coverageRate)) + "%",
	}
	fmt.Println(testInfo)
}

type TestInfo struct {
	Service      string
	CoverageRate string
	SuccessRate  string
	Method       Method
}
type Method struct {
	IsSuccess    string
	ErrorMsg     string
	ResponseTime int
}
