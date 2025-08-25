package fun

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
)

type Fun struct {
	connList    *sync.Map
	openFunc    func(id string)
	closeFunc   func(id string)
	serviceList map[string]*service
	boxList     *sync.Map
	guardList   []*any
	mu          sync.Mutex
	validate    *validator.Validate
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

func BindValidate(tag string, fn validator.Func) {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
			os.Exit(0)
		}
	}()
	f := GetFun()
	err := f.validate.RegisterValidation(tag, fn)
	if err != nil {
		panic(err.Error())
	}
}

func GetFun() *Fun {
	once.Do(func() {
		fun = &Fun{
			connList:    &sync.Map{},
			boxList:     &sync.Map{},
			serviceList: map[string]*service{},
			guardList:   []*any{},
			validate:    validator.New(),
		}
	})
	return fun
}

func Start(addr ...uint16) {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
		}
	}()
	http.HandleFunc("/", handleWebSocket(GetFun()))
	InfoLogger("Server started on port " + isPort(addr))
	err := http.ListenAndServe("0.0.0.0:"+isPort(addr), nil)
	if err != nil {
		panic(err.Error())
	}
}

func Gen() {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
		}
	}()
	err := os.RemoveAll(directory)
	if err != nil && !os.IsNotExist(err) {
		panic(err.Error())
	}
	genDefaultService()
}

func StartTls(certFile string, keyFile string, addr ...uint16) {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
		}
	}()
	http.HandleFunc("/", handleWebSocket(GetFun()))
	InfoLogger("Server started on port " + isPort(addr))
	err := http.ListenAndServeTLS("0.0.0.0:"+isPort(addr), certFile, keyFile, nil)
	if err != nil {
		panic(err.Error())
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
		if on.(onType).callBack != nil {
			callback := *on.(onType).callBack
			callback()
		}
		loadConnInfo.onList.Delete(requestId)
	}
}

func (fun *Fun) callGuard(service *service, ctx *Ctx) {
	var guardList []*any
	guardList = append(guardList, fun.guardList...)
	guardList = append(guardList, service.guardList...)
	for i := 0; i < len(guardList); i++ {
		guard := *guardList[i]
		g := guard.(Guard)
		g.Guard(*ctx)
	}
}

func (fun *Fun) cellMethod(ctx *Ctx, service *service, registeredMethod *method, requestData *reflect.Value, requestInfo *RequestInfo[map[string]any]) {
	fun.callGuard(service, ctx)
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
		//保存回调
		if connInfo, ok := fun.connList.Load(ctx.Id); ok {
			loadConnInfo := connInfo.(connInfoType)
			loadConnInfo.onList.Store(ctx.RequestId, onType{
				requestInfo.ServiceName,
				requestInfo.MethodName,
				nil,
			})
		}
		watchClose := func(callback func()) {
			if connInfo, ok := fun.connList.Load(ctx.Id); ok {
				loadConnInfo := connInfo.(connInfoType)
				loadConnInfo.onList.Store(ctx.RequestId, onType{
					requestInfo.ServiceName,
					requestInfo.MethodName,
					&callback,
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
	if !registeredMethod.isProxy || !value[0].IsNil() {
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
			if on.(onType).callBack != nil {
				callback := *on.(onType).callBack
				callback()
			}
			return true
		})
		fun.connList.Delete(id)
		if fun.closeFunc != nil {
			fun.closeFunc(id)
		}
	}
}

func BindService(service any, guardList ...Guard) {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
			os.Exit(0)
		}
	}()
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
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
			os.Exit(0)
		}
	}()
	f := GetFun()
	checkGuard(guard)
	guardWired(guard, f)
}

func (fun *Fun) returnData(id string, requestId string, data any, stackTrace string) {
	var result Result[any]
	// 尝试将 data 断言为 Result 类型
	if value, ok := data.(Result[any]); ok {
		result = value
		result.Id = requestId
		InfoLogger(result)
	} else {
		result = callError(getErrorString(data))
		result.Id = requestId
		ErrorLogger(getErrorString(data) + "\n" + stackTrace)
	}
	map1 := ToLowerMap(result)
	fun.send(id, map1)
}

func getErrorString(data any) string {
	if err, ok := data.(error); ok {
		return err.Error()
	} else {
		return fmt.Sprintf("%v", data)
	}
}

func ToLowerMap(obj interface{}) map[string]interface{} {
	// 先将对象序列化为 JSON 字节
	jsonBytes, _ := json.Marshal(obj)

	// 再将 JSON 字节反序列化为 map
	var m map[string]interface{}
	_ = json.Unmarshal(jsonBytes, &m)

	// 将所有键名转为小写
	return convertKeyToLowerCase(m)
}

// convertKeyToLowerCase 递归地将 map 中的所有键转换为小写
func convertKeyToLowerCase(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m {
		lowerKey := firstLetterToLower(k)

		// 如果值是 map 类型，递归处理
		if nestedMap, ok := v.(map[string]interface{}); ok {
			result[lowerKey] = convertKeyToLowerCase(nestedMap)
			// 如果值是数组类型，检查数组中的每个元素
		} else if arr, ok := v.([]interface{}); ok {
			result[lowerKey] = convertArrayKeyToLowerCase(arr)
		} else {
			result[lowerKey] = v
		}
	}

	return result
}

// convertArrayKeyToLowerCase 处理数组中的元素，如果元素是 map 则递归处理
func convertArrayKeyToLowerCase(arr []interface{}) []interface{} {
	result := make([]interface{}, len(arr))

	for i, item := range arr {
		if nestedMap, ok := item.(map[string]interface{}); ok {
			result[i] = convertKeyToLowerCase(nestedMap)
		} else if nestedArr, ok := item.([]interface{}); ok {
			result[i] = convertArrayKeyToLowerCase(nestedArr)
		} else {
			result[i] = item
		}
	}

	return result
}
