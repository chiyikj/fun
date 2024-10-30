package fun

import (
	"reflect"
)

// 检查bind参数是否合法
func checkParameter(methodType reflect.Type, methodName string, method *method, fun *Fun) {
	if methodType.NumIn() > 3 {
		panic("fun: The service " + methodName + " There can be only two parameters")
	}
	if methodType.NumIn() == 2 && (methodType.In(1).Kind() != reflect.Struct && methodType.In(1).Kind() != reflect.TypeOf(WatchClose(nil)).Kind()) {
		panic("fun: The service " + methodName + " In the case of one parameter it must be a structure or WatchClose")
	}
	if methodType.NumIn() == 3 && (methodType.In(1).Kind() != reflect.Struct && methodType.In(2).Kind() != reflect.TypeOf(WatchClose(nil)).Kind()) {
		panic("fun: The service " + methodName + " In the case of two arguments, the first argument must be a struct and the second must be WatchClose")
	}
	if methodType.In(1).Kind() == reflect.Struct {
		IsJsonType(methodType.In(1), fun)
		dto := methodType.In(1)
		method.dto = &dto
	}
}

// 检查返回值
func checkReturn(methodType reflect.Type, methodName string, method *method) {
	isProxy := (methodType.NumIn() == 2 && methodType.In(1).Kind() == reflect.TypeOf(WatchClose(nil)).Kind()) ||
		(methodType.NumIn() == 3 && methodType.In(2).Kind() == reflect.TypeOf(WatchClose(nil)).Kind())
	if methodType.NumOut() > 1 {
		panic("fun: The service " + methodName + " The return value can be only one")
	}
	if isProxy && (methodType.NumOut() == 0 || methodType.Out(0).Kind() != reflect.Ptr) {
		panic("fun: The service " + methodName + " Is a listener that must have a return value and be of pointer type")
	}
	if methodType.NumOut() == 1 {
		IsJsonType(methodType.Out(0), nil)
		if isProxy {
			returnData := methodType.Out(0)
			method.onType = &returnData
		}
	}
}

func checkCtx(serviceType reflect.Type, fun *Fun) {
	for i := 0; i < serviceType.NumField(); i++ {
		field := serviceType.Field(i)
		if field.Anonymous {
			panic("fun: service Anonymous fields are not supported")
		}
		fieldType := field.Type
		_, exists := fun.targets[fieldType]
		if !exists {
			panic("fun: service The type is not in the target list")
		}
	}
}
