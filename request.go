package fun

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
	"reflect"
	"testing"
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

func GetRequestInfo(test *testing.T, service any, methodName string, dto any, state map[string]string) RequestInfo[any] {
	if methodName == "" {
		test.Fatalf("abc: methodName cannot be empty")
	}
	t := reflect.TypeOf(service)
	if t.Kind() != reflect.Struct {
		test.Fatalf("abc: service must be a struct")
	}
	// 可选：检查方法是否存在
	method, exists := t.MethodByName(methodName)
	if !exists {
		test.Fatalf("abc: service does not have method " + methodName)
	}
	requestInfo := RequestInfo[any]{}
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
