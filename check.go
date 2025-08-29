package fun

import (
	"reflect"
	"strings"
	"unicode"
)

func checkService(t reflect.Type, f *Fun) {
	if t.Kind() == reflect.Ptr || t.Kind() != reflect.Struct {
		panic("Fun: " + t.Name() + " Service It must not be a pointer but a structure")
	}
	if isPrivate(t.Name()) {
		panic("Fun:" + t.Name() + " cannot be Private")
	}
	boxList := map[reflect.Type]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if i == 0 {
			if f.Type != reflect.TypeOf(Ctx{}) {
				panic("fun:Field one can only be without pointer Ctx")
			}
			if !f.Anonymous {
				panic("Fun:Ctx must be anonymous")
			}
		} else {
			checkBox(f, boxList)
		}
	}
	f.serviceList[t.Name()] = &service{
		serviceType: t,
		guardList:   []*any{},
		methodList:  map[string]*method{},
	}
}

func checkGuard(guard Guard) {
	t := reflect.TypeOf(guard)
	if t.Kind() == reflect.Ptr || t.Kind() != reflect.Struct {
		panic("Fun: " + t.Name() + " It must not be a pointer but a structure")
	}
	for i := 0; i < t.NumField(); i++ {
		boxList := map[reflect.Type]bool{}
		checkBox(t.Field(i), boxList)
	}
}

// 验证方法
func checkMethod(t reflect.Type, f *Fun) {
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		mt := m.Type
		method := &method{}
		isProxy := checkParameter(mt, m.Name, method)
		method.isProxy = isProxy
		method.method = m
		checkReturn(mt, m.Name, isProxy)
		f.serviceList[t.Name()].methodList[m.Name] = method
	}
}

func checkParameter(t reflect.Type, name string, method *method) bool {
	isProxy := false
	if t.NumIn() > 3 {
		panic("Fun: " + name + " Method cannot have more than two parameters")
	}
	if t.NumIn() == 3 {
		if t.In(2) != reflect.TypeOf((ProxyClose)(nil)) {
			panic("Fun: " + name + " parameter 2 is not a Proxy and must be a pointer")
		}
		isProxy = true
		if t.In(1).Kind() != reflect.Struct {
			panic("Fun: " + name + " Parameter is not struct")
		}
		checkType(t.In(1))
		dtoType := t.In(1)
		method.dto = &dtoType
	}
	if t.NumIn() == 2 {
		if t.In(1) == reflect.TypeOf((ProxyClose)(nil)) {
			isProxy = true
		} else {
			if t.In(1).Kind() != reflect.Struct {
				panic("Fun: " + name + " Parameter is not struct")
			}
			checkType(t.In(1))
			dtoType := t.In(1)
			method.dto = &dtoType
		}
	}
	return isProxy
}

func checkReturn(t reflect.Type, name string, isProxy bool) {
	if t.NumOut() > 1 {
		panic("Fun: " + name + " Method cannot have more than one return value")
	}
	if t.NumOut() == 1 {
		if isProxy {
			if t.Out(0).Kind() != reflect.Ptr {
				panic("Fun:Proxy " + name + " Method return value must be a pointer")
			}
		}
		checkType(t.Out(0))
	}
}

func isPrivate(value string) bool {
	return !unicode.IsUpper([]rune(value)[0])
}

func checkBox(s reflect.StructField, boxList map[reflect.Type]bool) {
	if _, ok := boxList[s.Type]; ok {
		return
	}
	boxList[s.Type] = true
	if s.Anonymous {
		panic("Fun:" + s.Name + " cannot be Anonymous")
	}
	if s.Type.Kind() != reflect.Ptr || s.Type.Elem().Kind() != reflect.Struct {
		panic("Fun:" + s.Name + " Must be a pointer and a struct")
	}
	if isPrivate(s.Name) {
		panic("Fun:" + s.Name + " cannot be Private")
	}
	//判断New函数是不是空参数
	if newMethod, found := s.Type.MethodByName("New"); found {
		// NumIn()返回值包括接收者本身，所以无参数的方法NumIn()应该等于1
		// NumOut()检查返回值数量，应该等于0
		if newMethod.Type.NumIn() != 1 || newMethod.Type.NumOut() != 0 {
			panic("Fun:" + s.Name + " New method must have no parameters and no return values")
		}
	}
	for i := 0; i < s.Type.Elem().NumField(); i++ {
		f := s.Type.Elem().Field(i)
		fieldTag := newTag(f.Tag)
		// 检查是否有 "auto" 标签
		if _, isAuto := fieldTag.GetTag("auto"); isAuto {
			checkBox(f, boxList)
		}
	}
}

func checkType(t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if strings.Contains(t.String(), "{}") {
		panic("Fun: " + t.Name() + " generic types containing 'any' or interface{} are not supported")
	}
	switch t.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.String, reflect.Bool:
		displayEnumType := reflect.TypeOf((*displayEnum)(nil)).Elem()
		enumType := reflect.TypeOf((*enum)(nil)).Elem()
		if t.Kind() == reflect.Uint8 && (t.Implements(displayEnumType) || t.Implements(enumType)) && isPrivate(t.Name()) {
			panic("Fun:" + t.Name() + " cannot be Private")
		}
		if t.Kind() == reflect.Uint8 && t.Implements(displayEnumType) {
			statusValue := reflect.New(t).Elem()
			enumValue := statusValue.Interface().(displayEnum)
			if len(enumValue.DisplayNames()) != len(enumValue.Names()) {
				panic("Fun: " + t.Name() + " enum names and display names must be the same length")
			}
		}
	case reflect.Struct:
		if t.NumField() == 0 {
			panic("Fun: " + t.Name() + " must have at least one field")
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if isPrivate(f.Name) {
				panic("Fun:" + f.Name + " cannot be Private")
			}
			checkType(f.Type)
		}
	case reflect.Slice:
		checkType(t.Elem())
	default:
		panic("fun:Unsupported types " + t.Name())
	}
}

func checkDto(dto *reflect.Type, dtoMap any) {
	dtoType := *dto
	if dtoType.Kind() == reflect.Struct {
		for i := 0; i < dtoType.NumField(); i++ {
			f := dtoType.Field(i)
			value, ok := dtoMap.(map[string]any)[firstLetterToLower(f.Name)]
			if f.Type.Kind() != reflect.Ptr && (!ok || value == nil) {
				panic(callError("Fun:" + f.Name + " Dto must be a pointer or have a corresponding field in the map"))
			}
			t := f.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if (t.Kind() == reflect.Struct || t.Kind() == reflect.Slice) && value != nil {
				checkDto(&t, value)
			}
		}
	} else {
		list, _ := dtoMap.([]any)
		for _, value := range list {
			if dtoType.Elem().Kind() != reflect.Ptr && value == nil {
				panic(callError("Fun:" + dtoType.Elem().Name() + " Dto must be a pointer or have a corresponding field in the map"))
			}
			t := dtoType.Elem()
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if (t.Kind() == reflect.Struct || t.Kind() == reflect.Slice) && value != nil {
				checkDto(&t, value)
			}
		}
	}
}
