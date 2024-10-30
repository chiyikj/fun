package fun

import (
	"reflect"
	"strings"
	"unicode"
)

func IsJsonType(t reflect.Type, fun *Fun) {
	basicTypes := map[reflect.Kind]struct{}{
		reflect.Int:    {},
		reflect.Int8:   {},
		reflect.Int16:  {},
		reflect.Int32:  {},
		reflect.Int64:  {},
		reflect.Uint:   {},
		reflect.Uint8:  {},
		reflect.Uint16: {},
		reflect.Uint32: {},
		reflect.Uint64: {},
		reflect.String: {},
		reflect.Bool:   {},
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Slice:
		elemType := t.Elem()
		if elemType.Kind() == reflect.Slice {
			// 不支持多维数组
			panic("fun:Two-dimensional arrays are not supported")
		}
		IsJsonType(elemType, nil)
	case reflect.Struct:
		if !unicode.IsUpper(rune(t.Name()[0])) {
			// 字段名不是首字母大写，不符合条件
			panic("fun:" + t.Name() + " Must be public")
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldType := field.Type
			//检查枚举

			if fun != nil {
				tag := field.Tag.Get("fun")
				if field.Tag.Get("fun") != "" {
					parts := strings.Split(tag, ",")
					for _, part := range parts {
						kv := strings.Split(part, ":")
						if fun.checkList[kv[0]] == nil {
							//枚举不存在
							panic("fun:" + kv[0] + " The authentication rule does not exist")
						}
					}
				}
			}
			// 检查字段名称的首字母是否大写
			if unicode.IsUpper(rune(field.Name[0])) {
				IsJsonType(fieldType, nil)
			} else {
				// 字段名不是首字母大写，不符合条件
				panic("fun:" + field.Name + " Must be public")
			}
		}
	default:
		_, exists := basicTypes[t.Kind()]
		if !exists {
			panic("fun:Unsupported types " + t.Name())
		}

	}
}

func isMapToStruct(dto reflect.Type, value1 reflect.Value, _map map[string]any, fun *Fun) {
	for i := 0; i < dto.NumField(); i++ {
		field := dto.Field(i)
		fieldType := field.Type
		var t = fieldType
		value, ok := _map[field.Name]
		if !ok {
			panic("fun:" + field.Name + " not  found")
		}
		if fieldType.Kind() != reflect.Ptr {
			if value == nil {
				panic("fun:" + field.Name + " not  found")
			}
		} else {
			t = t.Elem()
		}
		if value != nil {
			if t.Kind() == reflect.Slice {
				//目标类型
				elemType := t.Elem()
				t = fieldType
				if elemType.Kind() != reflect.Ptr {
					t = t.Elem()
				}
				mapValue, ok := _map[field.Name]
				if ok {
					sliceValue := mapValue.([]interface{})
					for _, item := range sliceValue {
						//非指针类型却传了一个空
						if elemType.Kind() != reflect.Ptr && sliceValue == nil {
							panic("fun:" + field.Name + " Non-pointer type with a nil value. This is not allowed.")
						}
						if t.Kind() == reflect.Struct {
							itemMap := item.(map[string]interface{})
							isMapToStruct(t, value1, itemMap, fun)
						}
					}
				}
			} else if t.Kind() == reflect.Struct {
				structMap := value.(map[string]interface{})
				isMapToStruct(t, value1, structMap, fun)
			}
		}
		tag := field.Tag.Get("fun")
		parts := strings.Split(tag, ",")
		for _, part := range parts {
			kv := strings.Split(part, ":")
			value := fun.checkList[kv[0]](fieldType, value1.Field(i).Interface(), kv[1])
			if value != nil {
				panic(value)
			}
		}
	}
}
