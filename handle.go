package fun

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type connInfoType struct {
	conn   *websocket.Conn
	mu     *sync.Mutex
	onList *sync.Map
}

func handleWebSocket(fun *Fun) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			return
		}
		//升级ws请求
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		var timer *time.Timer
		//到期回收
		defer func() {
			fun.closeFuncCell(&timer, conn, id)
		}()
		//客户端连接通知
		if fun.openFunc != nil {
			fun.openFunc(id)
		}
		fun.resetTimer(&timer, conn, id)
		//丢入客户端连接池
		fun.connList.Store(id, connInfoType{conn: conn, mu: &sync.Mutex{}, onList: &sync.Map{}})
		//封装用户上下文
		ctx := Ctx{Id: id, fun: fun}
		ctx.Ip = getIP(r)
		ctx.Push = func(id string, requestId string, data any) bool {
			return fun.Push(id, requestId, data)
		}
		ctx.Close = func(id string, requestId string) {
			fun.send(id, closeError(requestId))
			fun.close(id, requestId)
		}
		for {
			if fun.handleResponse(conn, &timer, ctx) {
				return
			}
		}
	}
}

// 处理请求
func (fun *Fun) handleResponse(conn *websocket.Conn, timer **time.Timer, ctx Ctx) bool {
	//获取信息
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return true
	}
	fun.handleMessage(messageType, &message, timer, conn, &ctx)
	return false
}

// 处理消息
func (fun *Fun) handleMessage(messageType int, message *[]byte, timer **time.Timer, conn *websocket.Conn, ctx *Ctx) {
	if messageType == websocket.BinaryMessage {
		//处理客户端ping信息 回复
		if len(*message) == 1 && (*message)[0] == 0 {
			fun.sendPong(ctx.Id)
			fun.resetTimer(timer, conn, ctx.Id)
		}
		return
	}
	InfoLogger(string(*message))
	//处理文本信息
	if messageType == websocket.TextMessage {
		var request RequestInfo[map[string]any]
		err := json.Unmarshal(*message, &request)
		defer func() {
			if err := recover(); err != nil {
				stackBuf := make([]byte, 8192)
				stackSize := runtime.Stack(stackBuf, false)
				stackTrace := string(stackBuf[:stackSize])
				fun.returnData(ctx.Id, request.Id, err, stackTrace)
			}
		}()
		if err != nil {
			panic(callError(err.Error()))
		}
		request.MethodName = firstLetterToUpper(request.MethodName)
		request.ServiceName = firstLetterToUpper(request.ServiceName)
		if request.Id == "" || request.MethodName == "" || request.ServiceName == "" {
			//处理为空的情况
			panic(callError("fun: request fields cannot be empty (id, methodName, serviceName)"))
		}
		ctx.RequestId = request.Id
		ctx.ServiceName = request.ServiceName
		ctx.MethodName = request.MethodName
		ctx.State = request.State
		fun.handleRequest(&request, ctx)
	}
}

// 处理请求
func (fun *Fun) handleRequest(request *RequestInfo[map[string]any], ctx *Ctx) {
	if request.Type == CloseType {
		fun.close(ctx.Id, ctx.RequestId)
	} else {
		fun.dto(request, ctx)
	}
}

// 处理参数
func (fun *Fun) dto(request *RequestInfo[map[string]any], ctx *Ctx) {

	service, ok := fun.serviceList[request.ServiceName]
	if !ok {
		panic(callError(fmt.Sprintf("fun: service %s not found", request.MethodName)))
	}
	method, ok := service.methodList[request.MethodName]
	if !ok {
		panic(callError(fmt.Sprintf("fun: method %s not found", request.MethodName)))
	}

	if request.Type != ProxyType && request.Type != FuncType {
		panic(callError(fmt.Sprintf("fun: method %s It is neither a function type nor a listener type", request.MethodName)))
	}

	if method.isProxy && request.Type != ProxyType {
		panic(callError(fmt.Sprintf("fun: method %s is a proxy but type mismatch", request.MethodName)))
	}

	if !method.isProxy && request.Type != FuncType {
		panic(callError(fmt.Sprintf("fun: method %s is not a proxy but type mismatch", request.MethodName)))
	}

	if method.dto != nil {
		if request.Dto == nil {
			panic(callError(fmt.Sprintf("fun: method %s requires a DTO but none provided", request.MethodName)))
		}
		// 创建目标类型的实例
		dto := reflect.New(*method.dto).Interface()

		// 将请求的DTO转换为JSON
		jsonData, err := json.Marshal(request.Dto)
		if err != nil {
			panic(callError(err.Error()))
		}

		// 反序列化到目标类型
		if err := json.Unmarshal(jsonData, dto); err != nil {
			panic(callError(err.Error()))
		}
		checkDto(method.dto, *request.Dto)
		err = fun.validate.Struct(dto)
		if err != nil {
			var err1 validator.ValidationErrors
			errors.As(err, &err1)
			panic(callError(err1[0].Error()))
		}
		requestData := reflect.ValueOf(dto).Elem()
		fun.cellMethod(ctx, service, method, &requestData, request)
	} else {
		fun.cellMethod(ctx, service, method, nil, request)
	}
}

func (fun *Fun) resetTimer(timer **time.Timer, conn *websocket.Conn, id string) {
	if *timer != nil {
		(*timer).Stop()
	}
	*timer = time.AfterFunc(7*time.Second, func() {
		fun.closeFuncCell(timer, conn, id)
	})
}
