package fun

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chiyikj/fun/util"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type Fun struct {
	//注入列表
	targets map[reflect.Type]reflect.Value
	//方法列表
	methods map[string]method
	//连接成功回调
	openFunc func(id string)
	//连接关闭回调
	closeFunc func(id string)
	//维护连接
	connList *sync.Map
	//自定义数据效验规则
	checkList map[string]checkFunc
}

const (
	function = iota
	proxy
)

type request struct {
	Id         string
	MethodName string
	Dto        any
	state      map[string]any
	MethodType int8
}

type method struct {
	//参数
	dto        *reflect.Type
	super      reflect.Type
	intercepts []interceptFunc
	onType     *reflect.Type
}

type ws struct {
	conn   *websocket.Conn
	mu     *sync.Mutex
	onList *sync.Map
}

type interceptFunc func(state map[string]any) *Result
type checkFunc func(p reflect.Type, value any, rule string) *Result

func New() *Fun {
	return &Fun{
		targets: map[reflect.Type]reflect.Value{
			reflect.TypeOf(Ctx{}): reflect.ValueOf(struct{}{}),
		},
		methods:   make(map[string]method),
		connList:  &sync.Map{},
		checkList: make(map[string]checkFunc),
	}
}

// Start 启动
func (fun *Fun) Start(addr uint16) {
	http.HandleFunc("/", handleWebSocket(fun))
	err := http.ListenAndServe(":"+strconv.Itoa(int(addr)), nil)
	fmt.Println(err)
}

// StartTls ssl启动
func (fun *Fun) StartTls(addr uint16, certFile string, keyFile string) {
	http.HandleFunc("/", handleWebSocket(fun))
	err := http.ListenAndServeTLS(":"+strconv.Itoa(int(addr)), certFile, keyFile, nil)
	fmt.Println(err)
}

// Check 绑定数据校验规则
func (fun *Fun) Check(key string, checkFunc checkFunc) {
	fun.checkList[key] = checkFunc
}

// Bind 绑定服务
func (fun *Fun) Bind(service any, intercepts ...interceptFunc) {
	serviceType := reflect.TypeOf(service)
	if serviceType.Kind() == reflect.Struct {
		if !unicode.IsUpper(rune(serviceType.Name()[0])) {
			// 字段名不是首字母大写，不符合条件
			panic("fun:" + serviceType.Name() + " Must be public")
		}
		//判断结构体属性是否合法
		checkCtx(serviceType, fun)
		// 判断结构体方法是否合法
		for i := 0; i < serviceType.NumMethod(); i++ {
			_method := serviceType.Method(i)
			var methodName = serviceType.Name() + "." + _method.Name
			// 获取方法的类型
			method := method{
				super: serviceType,
			}
			method.intercepts = intercepts
			//检查参数
			checkParameter(_method.Type, methodName, &method, fun)
			//检查返回值
			checkReturn(_method.Type, methodName, &method)
			fun.methods[methodName] = method
		}
	} else {
		panic("fun:service is not struct")
	}
}

// OnOpen 监听连接到服务器
func (fun *Fun) OnOpen(callback func(id string)) {
	fun.openFunc = callback
}

// OnClose OnOpen 监听关闭到服务器
func (fun *Fun) OnClose(callback func(id string)) {
	fun.closeFunc = callback
}

// Inject 依赖注入
func (fun *Fun) Inject(target any) {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("fun:target is not struct and Must be public")
	}
	if !unicode.IsUpper(rune(t.Name()[0])) {
		// 字段名不是首字母大写，不符合条件
		panic("fun:" + reflect.TypeOf(target).Name() + " Must be public")
	}
	fun.targets[reflect.TypeOf(target)] = reflect.ValueOf(target)
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (fun *Fun) closeFuncCell(timer **time.Timer, conn *websocket.Conn, id string) {
	conn.Close()
	if conn != nil && id != "" {
		if *timer != nil {
			(*timer).Stop()
		}
		if fun.closeFunc != nil {
			fun.closeFunc(id)
		}
		_conn, ok := fun.connList.Load(id)
		if ok {
			return
		}
		_conn.(ws).onList.Range(func(key, value any) bool {
			if value.(onType).callBack != nil {
				(*value.(onType).callBack)()
			}
			return true
		})
		fun.connList.Delete(id)
	}
}

func (fun *Fun) resetTimer(timer **time.Timer, conn *websocket.Conn) {
	if *timer != nil {
		(*timer).Stop()
	}
	*timer = time.AfterFunc(7*time.Second, func() {
		conn.Close()
	})
}

func handleWebSocket(fun *Fun) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		conn, err := upgrade.Upgrade(w, r, nil)
		var timer *time.Timer
		defer func() {
			fun.closeFuncCell(&timer, conn, id)
		}()
		if err != nil || id == "" {
			return
		}
		if fun.openFunc != nil {
			fun.openFunc(id)
		}
		fun.resetTimer(&timer, conn)
		fun.connList.Store(id, ws{conn: conn, mu: &sync.Mutex{}})
		ctx := Ctx{Ip: util.GetIp(r), Id: id, Send: fun.sender, Close: fun.close, fun: fun}
		for {
			if fun.handleWebSocketResponse(conn, &timer, id, ctx) {
				return
			}
		}
	}
}

func (fun *Fun) cellMethod(ctx Ctx, method method, data *reflect.Value, request *request) {
	cInstance := reflect.New(method.super).Elem()
	fun.inject(&cInstance, ctx)
	fun.interceptor(method, request)
	methodValue := cInstance.Addr().MethodByName(strings.Split(request.MethodName, ".")[1])
	var result Result
	var argumentsList []reflect.Value
	if method.dto != nil {
		argumentsList = append(argumentsList, data.Elem())
	}
	if method.onType != nil {
		argumentsList = append(argumentsList, reflect.ValueOf(ctx.Close))
	}
	values := methodValue.Call(argumentsList)
	if len(values) == 0 {
		result = resultSuccess(nil)
	} else {
		result = resultSuccess(values[0].Interface())
	}
	if method.onType == nil || result.Data != nil {
		panic(result)
	}
}

func (fun *Fun) send(id string, text any) bool {
	if conn, ok := fun.connList.Load(id); ok {
		ws := conn.(ws)
		ws.mu.Lock()
		err := ws.conn.WriteJSON(text)
		ws.mu.Unlock()
		return err == nil
	}
	return false
}

func (fun *Fun) returnData(request *request, id string, err any) {
	var result Result
	if _err, ok := err.(Result); ok {
		result = _err
	} else {
		result = resultCallError(err)
	}
	result.Id = request.Id
	fun.send(id, result)
}

func (fun *Fun) handleWebSocketResponse(conn *websocket.Conn, timer **time.Timer, id string, ctx Ctx) bool {
	var request = &request{}
	err := conn.ReadJSON(request)
	//统一处理错误
	defer func() {
		if err := recover(); err != nil {
			fun.returnData(request, id, err)
		}
	}()
	if err != nil {
		var syntaxError *json.SyntaxError
		var UnmarshalTypeError *json.UnmarshalTypeError
		if !errors.As(err, &syntaxError) && !errors.As(err, &UnmarshalTypeError) {
			return true
		}
		panic(err.Error())
	}
	fun.handleWebSocketRequest(timer, id, ctx, request, conn)
	return false
}

func (fun *Fun) handleWebSocketRequest(timer **time.Timer, id string, ctx Ctx, request *request, conn *websocket.Conn) {
	if request.MethodName == "ping" {
		fun.resetTimer(timer, conn)
		fun.send(id, Result{Data: "pong"})
	} else if request.MethodName == "close" {
		requestIdList := strings.Split(request.Id, ",")
		for _, requestId := range requestIdList {
			fun.close(id, requestId)
		}
	} else if request.Id == "" || request.MethodName == "" {
		panic("json: cannot unmarshal number into Go value of type cyi.Request")
	} else {
		fun.handleOtherRequests(ctx, request)
	}
}

func (fun *Fun) handleOtherRequests(ctx Ctx, request *request) {
	//校验参数
	method, ok := fun.methods[request.MethodName]
	if !ok {
		panic("fun: method not found")
	}
	//请求的类型不是一个监听者模式
	if request.MethodType == proxy && method.onType == nil {
		panic("fun: The type of request is not a listener pattern")
	}
	ctx.State = request.state
	ctx.MethodName = request.MethodName
	ctx.RequestId = request.Id
	if method.onType != nil {
		_conn, ok := fun.connList.Load(ctx.Id)
		if ok {
			_conn.(ws).onList.Store(request.Id, &onType{
				methodName: request.MethodName,
			})
		}
	}
	if method.dto != nil {
		marshal, err := json.Marshal(request.Dto)
		if err != nil {
			panic(err.Error())
		}
		newStruct := reflect.New(*method.dto)
		err = json.Unmarshal(marshal, newStruct.Interface())
		if err != nil {
			panic(err.Error())
		}
		dto, err1 := request.Dto.(map[string]any)
		if !err1 {
			panic("fun: A dto is not a map")
		}
		isMapToStruct(*method.dto, newStruct, dto, fun)
		fun.cellMethod(ctx, method, &newStruct, request)
	} else {
		if request.Dto != nil {
			panic("fun: Redundant parameter")
		}
		fun.cellMethod(ctx, method, nil, request)
	}
}

func (fun *Fun) close(id string, requestId string) {
	_conn, ok := fun.connList.Load(id)
	if !ok {
		return
	}
	on, ok := _conn.(ws).onList.Load(requestId)
	if ok {
		if on.(onType).callBack != nil {
			(*on.(onType).callBack)()
		}
		_conn.(ws).onList.Delete(requestId)
		fun.send(id, Result{
			Id:     requestId,
			Status: CloseError,
		})
	}
}

func (fun *Fun) sender(id string, requestId string, data any) bool {
	_conn, ok := fun.connList.Load(id)
	if !ok {
		return false
	}
	on, ok := _conn.(ws).onList.Load(requestId)
	if ok && reflect.TypeOf(data) == *fun.methods[on.(onType).methodName].onType {
		result := resultSuccess(data)
		result.Id = requestId
		return fun.send(id, result)
	}
	return false
}

func (fun *Fun) inject(cInstance *reflect.Value, ctx Ctx) {
	for i := 0; i < cInstance.NumField(); i++ {
		statsType := cInstance.Field(i).Type()
		if (statsType == reflect.TypeOf(Ctx{})) {
			cInstance.Field(i).Set(reflect.ValueOf(ctx))
		} else {
			cInstance.Field(i).Set(fun.targets[cInstance.Field(i).Type()])
		}
	}
}

func (fun *Fun) interceptor(method method, request *request) {
	for _, interceptor := range method.intercepts {
		if interceptor(request.state) == nil {
			break
		} else {
			panic(interceptor(request.state))
		}
	}
}
