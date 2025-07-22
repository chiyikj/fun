package fun

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
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

func GetRequestInfo(dto any, state map[string]string) RequestInfo[any] {
	serviceOnMethodByName := getServiceOnMethodByName()
	id, _ := gonanoid.New()
	return RequestInfo[any]{
		Id:          id,
		ServiceName: serviceOnMethodByName.serviceName,
		MethodName:  serviceOnMethodByName.methodName,
		Dto:         &dto,
		State:       state,
		Type:        FuncType,
	}
}
