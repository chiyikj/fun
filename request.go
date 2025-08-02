package fun

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
	"reflect"
)

const (
	FuncType = iota
	ProxyType
	CloseType
)

type RequestInfo[T any] struct {
	Id          string
	MethodName  string
	ServiceName string
	Dto         *T
	State       map[string]string
	Type        uint8
}

func GetRequestInfo[T any](service any, methodName string, dto T, state map[string]string) RequestInfo[T] {
	if methodName == "" {
		panic("test: methodName cannot be empty")
	}
	t := reflect.TypeOf(service)
	if t.Kind() != reflect.Struct {
		panic("test: service must be a struct")
	}
	// 可选：检查方法是否存在
	method, exists := t.MethodByName(methodName)
	if !exists {
		panic("test: service does not have method " + methodName)
	}
	var requestInfo RequestInfo[T] = RequestInfo[T]{}
	if method.Type.In(method.Type.NumIn()-1) == reflect.TypeOf((ProxyClose)(nil)) {
		requestInfo.Type = ProxyType
	} else {
		requestInfo.Type = FuncType
	}
	id, _ := gonanoid.New()
	requestInfo.Id = id
	requestInfo.ServiceName = t.Name()
	requestInfo.MethodName = methodName
	requestInfo.Dto = &dto
	requestInfo.State = state
	return requestInfo
}
