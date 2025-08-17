package fun

import (
	"os"
	"reflect"
	"runtime"
)

func boxWired(data any, fun *Fun) {
	serviceInstance := reflect.New(reflect.TypeOf(data)).Elem()
	// 注入结构体字段依赖
	for i := 0; i < serviceInstance.NumField(); i++ {
		field := serviceInstance.Field(i)
		if field.Type() != reflect.TypeOf(Ctx{}) {
			if _, isWired := fun.boxList.Load(field.Type()); !isWired {
				fun.autowired(field)
			}
		}
	}
}

func serviceGuardWired(serviceName string, data Guard, fun *Fun) {
	guardInstance := reflect.New(reflect.TypeOf(data)).Elem()
	for i := 0; i < guardInstance.NumField(); i++ {
		field := guardInstance.Field(i)
		if box, isWired := fun.boxList.Load(field.Type()); isWired {
			field.Set(box.(reflect.Value))
		} else {
			fun.autowired(field)
		}
	}
	guard := guardInstance.Interface()
	fun.serviceList[serviceName].guardList = append(fun.serviceList[serviceName].guardList, &guard)
}

func guardWired(data Guard, fun *Fun) {
	guardInstance := reflect.New(reflect.TypeOf(data)).Elem()
	for i := 0; i < guardInstance.NumField(); i++ {
		field := guardInstance.Field(i)
		if box, isWired := fun.boxList.Load(field.Type()); isWired {
			field.Set(box.(reflect.Value))
		} else {
			fun.autowired(field)
		}
	}
	guard := guardInstance.Interface()
	fun.guardList = append(fun.guardList, &guard)
}

func Wired[T any]() *T {
	defer func() {
		if err := recover(); err != nil {
			stackBuf := make([]byte, 8192)
			stackSize := runtime.Stack(stackBuf, false)
			stackTrace := string(stackBuf[:stackSize])
			PanicLogger(getErrorString(err) + "\n" + stackTrace)
			os.Exit(0)
		}
	}()
	var data1 T
	t := reflect.TypeOf(data1)
	data := new(T)
	if t.Kind() != reflect.Struct {
		panic("Fun: " + t.Name() + " It must be a structure")
	}
	if isPrivate(t.Name()) {
		panic("Fun:" + t.Name() + " cannot be Private")
	}
	GetFun()
	fun.mu.Lock()
	t1 := reflect.TypeOf(data)
	if box, isWired := fun.boxList.Load(t1); isWired {
		fun.mu.Unlock()
		return box.(reflect.Value).Interface().(*T)
	}
	v := reflect.ValueOf(data)
	fun.boxList.Store(t1, v)
	fun.mu.Unlock()
	boxList := map[reflect.Type]bool{}
	for i := 0; i < t.NumField(); i++ {
		c := t.Field(i)
		fieldTag := newTag(c.Tag)

		// 检查是否有 "auto" 标签
		if _, isAuto := fieldTag.GetTag("auto"); isAuto {
			// 查找是否已有该类型的依赖
			if dependency, loaded := fun.boxList.Load(c.Type); loaded {
				// 如果已存在，直接赋值
				v.Elem().Field(i).Set(dependency.(reflect.Value))
			} else {
				// 否则递归注入该字段
				checkBox(c, boxList)
				fun.autowired(v.Elem().Field(i))
			}
		}
	}
	// 调用 New() 方法（如果存在）
	newMethod := v.MethodByName("New")
	if newMethod.IsValid() {
		newMethod.Call(nil)
	}
	return data
}

func (fun *Fun) autowired(fieldValue reflect.Value) {
	// 创建当前字段类型的实例（指针类型）
	instance := reflect.New(fieldValue.Type().Elem())

	// 存储到依赖池中，供后续注入使用
	fun.boxList.Store(fieldValue.Type(), instance)

	fieldValue.Set(instance)
	// 获取实例的结构体值（解引用指针）
	structValue := instance.Elem()

	// 遍历结构体字段，处理带有 `auto` 标签的字段
	for i := 0; i < structValue.NumField(); i++ {
		structField := structValue.Type().Field(i)
		fieldTag := newTag(structField.Tag)

		// 检查是否有 "auto" 标签
		if _, isAuto := fieldTag.GetTag("auto"); isAuto {
			// 查找是否已有该类型的依赖
			if dependency, loaded := fun.boxList.Load(structField.Type); loaded {
				// 如果已存在，直接赋值
				structValue.Field(i).Set(dependency.(reflect.Value))
			} else {
				// 否则递归注入该字段
				fun.autowired(structValue.Field(i))
			}
		}
	}

	// 调用 New() 方法（如果存在）
	newMethod := instance.MethodByName("New")
	if newMethod.IsValid() {
		newMethod.Call(nil)
	}
}

func (fun *Fun) serviceWired(serviceInstance reflect.Value, ctx *Ctx) {
	// 注入结构体字段依赖
	for i := 0; i < serviceInstance.NumField(); i++ {
		field := serviceInstance.Field(i)
		// 如果字段是 Ctx 类型，则注入上下文
		if field.Type() == reflect.TypeOf(Ctx{}) {
			field.Set(reflect.ValueOf(*ctx))
		} else {
			dependency, _ := fun.boxList.Load(field.Type())
			field.Set(dependency.(reflect.Value))
		}
	}
}
