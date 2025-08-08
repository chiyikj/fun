package mock

import (
	"encoding/json"
	"github.com/chiyikj/fun"
	"reflect"
	"time"
)

var testPort *uint16 = nil

var testMessageQueue = make(chan []byte, 100)

func getMessage(id string, result any) {
	for {
		select {
		case message := <-testMessageQueue:

			// 创建一个临时结构体来解析ID
			var tempResult struct {
				Id string `json:"id"`
			}

			// 解析消息以获取ID
			err := json.Unmarshal(message, &tempResult)
			if err != nil {
				panic(err)
			}

			// 检查ID是否一致
			if tempResult.Id != id {
				break
			}
			// 将消息反序列化到目标结果中
			err = json.Unmarshal(message, result)
			if err != nil {
				break
			}
			return
		}
	}
}

func mockSendJson(requestInfo any) {
	writeMutex.Lock()
	err := testClient.WriteJSON(requestInfo)
	writeMutex.Unlock()
	if err != nil {
		panic(err)
	}
}

func MockRequest[T any](requestInfo any) fun.Result[T] {
	requestId := reflect.ValueOf(requestInfo).FieldByName("Id").String()
	mockSendJson(requestInfo)
	result := fun.Result[T]{}
	getMessage(requestId, &result)
	return result
}

type ProxyMessage struct {
	Message func(message any)
	Close   func()
}

func MockProxyClose(id string) {
	requestInfo := fun.RequestInfo[any]{
		Id:   id,
		Type: fun.CloseType,
	}
	mockSendJson(requestInfo)
}

func MockProxy(requestInfo any, proxy ProxyMessage, seconds int64) {
	requestId := reflect.ValueOf(requestInfo).FieldByName("Id").String()
	mockSendJson(requestInfo)
	GetProxyMessage(requestId, proxy, seconds)
}

func GetProxyMessage(id string, proxy ProxyMessage, seconds int64) {
	timeout := time.After(time.Duration(seconds) * time.Second)
	for {
		select {
		case message := <-testMessageQueue:

			// 创建一个临时结构体来解析ID
			var tempResult struct {
				Id string `json:"id"`
			}

			// 解析消息以获取ID
			err := json.Unmarshal(message, &tempResult)
			if err != nil {
				panic(err)
			}

			// 检查ID是否一致
			if tempResult.Id != id {
				break
			}

			var result = fun.Result[any]{}
			if result.Status == fun.CloseErrorCode {
				if proxy.Close != nil {
					proxy.Close()
				}
				return
			}

			// 将消息反序列化到目标结果中
			err = json.Unmarshal(message, &result)
			if err != nil {
				break
			}
			proxy.Message(result.Data)
		case <-timeout:
			mockSendJson(fun.RequestInfo[any]{
				Id:   id,
				Type: fun.CloseType,
			})
			return
		}
	}
}

func init() {
	port := randomPort()
	testPort = &port
	go func() {
		fun.Start(port)
	}()
	client(*testPort)
}
