package fun

import (
	"runtime"
	"strings"
)

// 获取调用函数名
func getServiceOnMethodByName() serviceOnMethodByName {
	pc, _, _, _ := runtime.Caller(2)
	funcObj := runtime.FuncForPC(pc)
	name := funcObj.Name()
	name = strings.Replace(name, "Test", "", -1)

	// 按 . 分割
	parts := strings.Split(name, ".")

	// 取最后两个部分：结构体类型 + 方法名
	lastTwo := parts[len(parts)-2:]
	return serviceOnMethodByName{
		lastTwo[0],
		lastTwo[1],
	}
}

type serviceOnMethodByName struct {
	serviceName string
	methodName  string
}
