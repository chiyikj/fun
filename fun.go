package fun

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"sync"
	"time"
)

type Fun struct {
	connList    *sync.Map
	openFunc    func(id string)
	closeFunc   func(id string)
	serviceList map[string]*service
	boxList     *sync.Map
	guardList   []*any
}

type service struct {
	serviceType reflect.Type
	guardList   []*any
	methodList  map[string]*method
}

type method struct {
	dto     *reflect.Type
	method  reflect.Method
	isProxy bool
}

var (
	once sync.Once
	fun  *Fun
)

func GetFun() *Fun {
	once.Do(func() {
		fun = &Fun{
			connList:    &sync.Map{},
			boxList:     &sync.Map{},
			serviceList: map[string]*service{},
			guardList:   []*any{},
		}
	})
	return fun
}

func Start(addr ...uint16) {
	http.HandleFunc("/", handleWebSocket(GetFun()))
	err := http.ListenAndServe("127.0.0.1:"+isPort(addr), nil)
	if err != nil {
		panic(err)
	}
}

func Gen() {
	genDefaultService()
}

func typeToJsType(t reflect.Type) string {
	text := ""
	if t.Kind() == reflect.Ptr {
		text += " | null"
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		text = "number" + text
		break
	case reflect.Bool:
		text = "boolean" + text
		break
	case
		reflect.String, reflect.Struct:
		text = t.Name() + text
		break
	default:
		text = typeToJsType(t.Elem()) + "[]" + text
		break
	}
	return text
}

func StartTls(certFile string, keyFile string, addr ...uint16) {
	http.HandleFunc("/", handleWebSocket(GetFun()))
	err := http.ListenAndServeTLS("localhost:"+isPort(addr), certFile, keyFile, nil)
	if err != nil {
		panic(err)
	}
}

func (fun *Fun) close(id string, requestId string) {
	connInfo, ok := fun.connList.Load(id)
	if !ok {
		return
	}
	loadConnInfo := connInfo.(connInfoType)
	on, ok := loadConnInfo.onList.Load(requestId)
	if ok {
		on.(onType).callBack()
		loadConnInfo.onList.Delete(requestId)
	}
}

func (fun *Fun) callGuard(service *service, serviceName string, methodName string, requestInfo *RequestInfo[map[string]any]) {
	var guardList []*any
	guardList = append(guardList, fun.guardList...)
	guardList = append(guardList, service.guardList...)
	for i := 0; i < len(guardList); i++ {
		guard := *guardList[i]
		g := guard.(Guard)
		if err := g.Guard(serviceName, methodName, requestInfo.State); err != nil {
			panic(*err)
		}
	}
}

func (fun *Fun) cellMethod(ctx *Ctx, service *service, registeredMethod *method, requestData *reflect.Value, requestInfo *RequestInfo[map[string]any]) {
	fun.callGuard(service, requestInfo.ServiceName, requestInfo.MethodName, requestInfo)
	// 创建目标方法所属结构体的实例
	serviceInstance := reflect.New(service.serviceType).Elem()

	methodValue := serviceInstance.Addr().MethodByName(requestInfo.MethodName)
	fun.serviceWired(serviceInstance, ctx)
	var result Result[any]
	var args []reflect.Value
	if requestData != nil {
		args = append(args, *requestData)
	}
	if registeredMethod.isProxy {
		watchClose := func(callback func()) {
			//保存回调
			if connInfo, ok := fun.connList.Load(ctx.Id); ok {
				loadConnInfo := connInfo.(connInfoType)

				loadConnInfo.onList.Store(ctx.RequestId, onType{
					requestInfo.ServiceName,
					requestInfo.MethodName,
					callback,
				})
			}
		}
		args = append(args, reflect.ValueOf(watchClose))
	}

	value := methodValue.Call(args)
	if len(value) == 0 {
		result = success(nil)
	} else {
		result = success(value[0].Interface())
	}
	if !registeredMethod.isProxy || result.Data != nil {
		panic(result)
	}

}

func (fun *Fun) closeFuncCell(timer **time.Timer, conn *websocket.Conn, id string) {
	_ = conn.Close()
	if conn != nil {
		if *timer != nil {
			(*timer).Stop()
		}
		connInfo, ok := fun.connList.Load(id)
		if !ok {
			return
		}
		connInfo.(connInfoType).onList.Range(func(_, on any) bool {
			on.(onType).callBack()
			return true
		})
		fun.connList.Delete(id)
		if fun.closeFunc != nil {
			fun.closeFunc(id)
		}
	}
}

func BindService(service any, guardList ...Guard) {
	f := GetFun()
	t := reflect.TypeOf(service)
	checkService(t, f)
	checkMethod(t, f)
	boxWired(service, f)
	for _, guard := range guardList {
		checkGuard(guard)
		serviceGuardWired(t.Name(), guard, f)
	}

}

func BindGuard(guard Guard) {
	f := GetFun()
	checkGuard(guard)
	guardWired(guard, f)
}

func (fun *Fun) returnData(id string, requestId string, data any) {
	var result Result[any]
	// 尝试将 data 断言为 Result 类型
	switch value := data.(type) {
	case Result[any]:
		result = value
	case error:
		result = callError(value.Error())
	default:
		result = callError(fmt.Sprintf("%v", data))
	}
	result.Id = requestId
	fun.send(id, result)
}
