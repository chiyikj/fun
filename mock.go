package fun

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var testPort *uint16 = nil

var testMessageQueue = make(chan []byte, 100)

func getMessage(t *testing.T, id string, result any) {
	timeout := time.After(10 * time.Second)
	for {
		select {
		case message := <-testMessageQueue:

			// 创建一个临时结构体来解析ID
			var tempResult struct {
				Id string `json:"id"`
			}

			// 解析消息以获取ID
			_ = json.Unmarshal(message, &tempResult)

			// 检查ID是否一致
			if tempResult.Id != id {
				break
			}
			// 将消息反序列化到目标结果中
			err := json.Unmarshal(message, result)
			if err != nil {
				ErrorLogger(fmt.Sprintf("%v", err))
				t.Fail()
			}
			return
		case <-timeout:
			ErrorLogger(callError("fun:request timeout"))
			t.Fail()
			return
		}
	}
}

func mockSendJson(t *testing.T, requestInfo any) {
	writeMutex.Lock()
	err := testClient.WriteJSON(requestInfo)
	writeMutex.Unlock()
	if err != nil {
		ErrorLogger(fmt.Sprintf("%v", err))
		t.Fail()
	}
}

func MockRequest[T any](t *testing.T, requestInfo any) Result[T] {
	newClientOrService()
	requestId := reflect.ValueOf(requestInfo).FieldByName("Id").String()
	mockSendJson(t, requestInfo)
	result := Result[T]{}
	getMessage(t, requestId, &result)
	result.Id = requestId
	return result
}

type ProxyMessage struct {
	Message func(message any)
	Close   func()
}

func MockProxyClose(t *testing.T, id string) {
	requestInfo := RequestInfo[any]{
		Id:   id,
		Type: CloseType,
	}
	mockSendJson(t, requestInfo)
}

func MockProxy(t *testing.T, requestInfo any, proxy ProxyMessage, seconds int64) {
	newClientOrService()
	requestId := reflect.ValueOf(requestInfo).FieldByName("Id").String()
	mockSendJson(t, requestInfo)
	GetProxyMessage(t, requestId, proxy, seconds)
}

func GetProxyMessage(t *testing.T, id string, proxy ProxyMessage, seconds int64) {
	timeout := time.After(time.Duration(seconds) * time.Second)
	for {
		select {
		case message := <-testMessageQueue:

			// 创建一个临时结构体来解析ID
			var tempResult struct {
				Id string `json:"id"`
			}

			// 解析消息以获取ID
			_ = json.Unmarshal(message, &tempResult)

			// 检查ID是否一致
			if tempResult.Id != id {
				break
			}

			var result = Result[any]{}
			if result.Status == closeErrorCode {
				if proxy.Close != nil {
					proxy.Close()
				}
				return
			}

			// 将消息反序列化到目标结果中
			err := json.Unmarshal(message, &result)
			if err != nil {
				break
			}
			proxy.Message(result.Data)
		case <-timeout:
			mockSendJson(t, RequestInfo[any]{
				Id:   id,
				Type: CloseType,
			})
			return
		}
	}
}

func newClientOrService() {
	if testPort == nil {
		port := randomPort()
		testPort = &port
		go func() {
			Start(port)
		}()
		client(*testPort)
	}
}
