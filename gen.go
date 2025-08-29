package fun

import (
	"bytes"
	"os"
	"reflect"
	"slices"
	"strings"
	"text/template"
	"unicode"
)

var directory = "../gen/ts/"

type gen struct {
	GenServiceList []*genServiceType
}

type genMethodType struct {
	MethodName      string
	ReturnValueText string
	DtoText         string
	ArgsText        string
	GenericTypeText string
}

type genEnumType struct {
	Names        []string
	displayNames []string
	Name         string
}

type genImportType struct {
	Name string
	Path string
}

type genServiceType struct {
	ServiceName       string
	GenMethodTypeList []*genMethodType
	GenImport         []*genImportType
	IsIncludeProxy    bool
}

type genClassType struct {
	Name              string
	GenImport         []*genImportType
	GenClassFieldType []*genClassFieldType
}

type genClassFieldType struct {
	Name string
	Type string
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
		if t.Kind() == reflect.Uint8 {
			enumType := reflect.TypeOf((*enum)(nil)).Elem()
			displayEnumType := reflect.TypeOf((*displayEnum)(nil)).Elem()
			if t.Implements(displayEnumType) || t.Implements(enumType) {
				text = t.Name() + text
			} else {
				text = "number" + text
			}
		} else {
			text = "number" + text
		}
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

func genService(
	service *service,
	serviceContext *genServiceType,
	visitedStructPaths []string,
	visitedEnumPaths []string,
) {
	// 收集服务方法中涉及的结构体导入路径
	var nestedImports []*genImportType

	// 遍历服务中的每个方法
	isIncludeProxy := false
	for _, method := range service.methodList {
		var returnValueText string
		var dtoText string
		var argsText string
		var genericTypeText string

		if method.isProxy {
			isIncludeProxy = true
		}
		if method.method.Type.NumOut() == 0 {
			returnValueText = "null"
		} else {
			returnType := method.method.Type.Out(0)

			// 转换为 TypeScript 类型
			// 创建结构体模板数据
			//处理泛类
			t := firstLetterToLower(typeToJsType(returnType))
			if !strings.Contains(t, "[]") && strings.Contains(t, "[") {
				returnValueText = getGenericTypeName(t) + parseGenericTypeParams(t)
			} else {
				returnValueText = t
			}

			// 如果是结构体类型，递归生成导入路径
			if returnType.Kind() == reflect.Ptr {
				returnType = returnType.Elem()
			}
			if returnType.Kind() == reflect.Struct {
				nestedImports = append(nestedImports, genStruct(returnType, visitedStructPaths, visitedEnumPaths))
			}
			if returnType.Kind() == reflect.Slice {
				fieldType := returnType.Elem()
				if fieldType.Kind() == reflect.Ptr {
					fieldType = fieldType.Elem()
				}
				if fieldType.Kind() == reflect.Struct {
					nestedImports = append(nestedImports, genStruct(fieldType, visitedStructPaths, visitedEnumPaths))
				}
			}
			enumType := reflect.TypeOf((*enum)(nil)).Elem()
			displayEnumType := reflect.TypeOf((*displayEnum)(nil)).Elem()
			if returnType.Kind() == reflect.Uint8 && (returnType.Implements(displayEnumType) || returnType.Implements(enumType)) {
				nestedImports = append(nestedImports, getEnum(returnType, visitedEnumPaths))
			}
		}

		// 处理 DTO 参数
		if method.dto != nil {
			v := firstLetterToLower(typeToJsType(*method.dto))
			if !strings.Contains(v, "[]") && strings.Contains(v, "[") {
				dtoText += "dto:" + getGenericTypeName(v) + parseGenericTypeParams(v)
			} else {
				dtoText += "dto:" + v
			}
			argsText += ",dto"
			nestedImports = append(nestedImports, genStruct(*method.dto, visitedStructPaths, visitedEnumPaths))
		}

		// 处理代理逻辑（on 回调）
		if method.isProxy {
			if method.dto != nil {
				dtoText += ","
			}
			dtoText += "on:on<" + strings.ReplaceAll(returnValueText, " | null", "") + ">"
			argsText += ",on"
			genericTypeText = strings.ReplaceAll(returnValueText, " | null", "")
			returnValueText = "() => void"
		} else {
			genericTypeText = returnValueText
			returnValueText = "result<" + returnValueText + ">"
		}

		// 添加方法信息到服务上下文
		serviceContext.GenMethodTypeList = append(serviceContext.GenMethodTypeList, &genMethodType{
			MethodName:      firstLetterToLower(method.method.Name),
			ReturnValueText: returnValueText,
			DtoText:         dtoText,
			ArgsText:        argsText,
			GenericTypeText: firstLetterToLower(genericTypeText),
		})
	}
	serviceContext.IsIncludeProxy = isIncludeProxy
	// 去重导入路径
	serviceContext.GenImport = deduplicateServiceImports(nestedImports)

	// 生成 TypeScript 文件
	genCode(
		genServiceTemplate(),
		"",
		firstLetterToLower(service.serviceType.Name()),
		serviceContext,
	)
}

func deduplicateServiceImports(imports []*genImportType) []*genImportType {
	seen := make(map[string]bool)
	var result []*genImportType

	for _, imp := range imports {
		if !seen[imp.Path] {
			seen[imp.Path] = true
			result = append(result, imp)
		}
	}

	return result
}

func genDefaultService() {
	f := GetFun()
	var visitedStructPaths []string
	var visitedEnumPaths []string
	genContext := gen{GenServiceList: []*genServiceType{}}

	for _, service := range f.serviceList {
		serviceContext := &genServiceType{
			ServiceName:       firstLetterToLower(service.serviceType.Name()),
			GenMethodTypeList: []*genMethodType{},
		}

		genContext.GenServiceList = append(genContext.GenServiceList, serviceContext)
		genService(service, serviceContext, visitedStructPaths, visitedEnumPaths)
	}

	genCode(genDefaultServiceTemplate(), "", "fun", genContext)
}

func parseGenericTypeParams(typeName string) string {
	// 查找第一个 '[' 的位置
	start := strings.Index(typeName, "[")
	// 查找最后一个 ']' 的位置
	end := strings.LastIndex(typeName, "]")

	// 提取中括号内的内容
	paramsStr := typeName[start+1 : end]

	// 简单处理，按逗号分割（不考虑嵌套情况）
	params := strings.Split(paramsStr, ",")

	// 去除空格
	for i, param := range params {
		LL := strings.Split(strings.TrimSpace(param), ".")
		params[i] = firstLetterToUpper(LL[len(LL)-1])
	}

	return strings.Join(params, "")
}

func getGenericTypeName(typeName string) string {
	// 查找第一个 '[' 的位置
	start := strings.Index(typeName, "[")

	// 提取中括号内的内容
	paramsStr := typeName[0:start]

	return paramsStr
}

func genStruct(t reflect.Type, visitedPaths []string, visitedEnumPaths []string) *genImportType {
	// 提取结构体所在的包路径并生成相对路径
	pkgParts := strings.Split(t.PkgPath(), "/")
	relativePath := strings.Join(pkgParts[1:], "/")
	// 如果路径已生成过，直接返回引用
	if slices.Contains(visitedPaths, relativePath) {
		return &genImportType{Name: firstLetterToLower(t.Name()), Path: relativePath}
	}

	var structTemplate genClassType
	// 创建结构体模板数据
	if !strings.Contains(t.String(), "[]") && strings.Contains(t.String(), "[") {
		structTemplate = genClassType{
			Name: firstLetterToLower(getGenericTypeName(t.Name())) + parseGenericTypeParams(t.Name()),
		}
	} else {
		structTemplate = genClassType{
			Name: firstLetterToLower(t.Name()),
		}
	}
	// 收集嵌套结构体的导入路径
	var nestedImports []*genImportType

	// 遍历结构体字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		jsType := typeToJsType(fieldType)
		name := field.Name
		// 解引用指针
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			name += "?"
		}
		// 生成字段类型并添加到模板
		if !strings.Contains(jsType, "[]") && strings.Contains(jsType, "[") {
			structTemplate.GenClassFieldType = append(structTemplate.GenClassFieldType, &genClassFieldType{
				Name: firstLetterToLower(name),
				Type: firstLetterToLower(getGenericTypeName(jsType)) + parseGenericTypeParams(jsType),
			})
		} else {
			structTemplate.GenClassFieldType = append(structTemplate.GenClassFieldType, &genClassFieldType{
				Name: firstLetterToLower(name),
				Type: firstLetterToLower(jsType),
			})
		}

		// 如果字段是结构体，递归生成导入路径
		if fieldType.Kind() == reflect.Struct {
			nestedImports = append(nestedImports, genStruct(fieldType, visitedPaths, visitedEnumPaths))
		}

		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Struct {
			nestedImports = append(nestedImports, genStruct(fieldType.Elem(), visitedPaths, visitedEnumPaths))
		}

		enumType := reflect.TypeOf((*enum)(nil)).Elem()
		displayEnumType := reflect.TypeOf((*displayEnum)(nil)).Elem()

		if fieldType.Kind() == reflect.Uint8 && (fieldType.Implements(displayEnumType) || fieldType.Implements(enumType)) {
			nestedImports = append(nestedImports, getEnum(fieldType, visitedEnumPaths))
		}

	}

	// 去重并计算相对路径
	basePath := strings.Split(relativePath, "/")
	uniqueImports := deduplicateStructImports(nestedImports, basePath)

	// 将去重后的导入路径添加到结构体模板中
	structTemplate.GenImport = uniqueImports

	// 生成 TypeScript 文件
	genCode(
		genStructTemplate(),
		relativePath,
		structTemplate.Name,
		structTemplate,
	)

	// 标记该路径已生成
	visitedPaths = append(visitedPaths, relativePath)
	if !strings.Contains(t.String(), "[]") && strings.Contains(t.String(), "[") {
		return &genImportType{
			Name: structTemplate.Name,
			Path: relativePath + "/" + structTemplate.Name,
		}
	} else {
		return &genImportType{
			Name: firstLetterToLower(t.Name()),
			Path: relativePath + "/" + structTemplate.Name,
		}
	}
}

func getEnum(t reflect.Type, visitedEnumPaths []string) *genImportType {
	pkgParts := strings.Split(t.PkgPath(), "/")
	relativePath := strings.Join(pkgParts[1:], "/")
	// 如果路径已生成过，直接返回引用
	if slices.Contains(visitedEnumPaths, relativePath) {
		return &genImportType{Name: firstLetterToLower(t.Name()), Path: relativePath}
	}

	var enumTemplate genEnumType
	displayEnumType := reflect.TypeOf((*displayEnum)(nil)).Elem()
	statusValue := reflect.New(t).Elem()
	if t.Implements(displayEnumType) {
		enumValue := statusValue.Interface().(displayEnum)
		enumTemplate.Names = enumValue.Names()
		enumTemplate.displayNames = enumValue.DisplayNames()
	} else {
		enumValue := statusValue.Interface().(enum)
		enumTemplate.Names = enumValue.Names()
	}

	genCode(
		genEnumTemplate(),
		relativePath,
		firstLetterToLower(t.Name()),
		enumTemplate,
	)

	return &genImportType{Name: firstLetterToLower(t.Name()), Path: relativePath}
}

func deduplicateStructImports(imports []*genImportType, basePath []string) []*genImportType {
	seen := make(map[string]bool)
	var result []*genImportType

	for _, imp := range imports {
		if seen[imp.Path] {
			continue
		}

		// 计算相对路径
		impPathParts := strings.Split(imp.Path, "/")
		impPathParts[len(impPathParts)-1] = firstLetterToLower(impPathParts[len(impPathParts)-1])
		commonPrefixLen := 0
		for i := 0; i < len(basePath) && i < len(impPathParts); i++ {
			if basePath[i] != impPathParts[i] {
				break
			}
			commonPrefixLen++
		}
		// 构建相对路径前缀
		var relativePathPrefix string
		for i := commonPrefixLen; i < len(basePath); i++ {
			relativePathPrefix += "../"
		}

		// 保存结果
		seen[imp.Path] = true
		result = append(result, &genImportType{
			Name: firstLetterToLower(imp.Name),
			Path: relativePathPrefix + strings.Join(impPathParts[commonPrefixLen:], "/"),
		})
	}

	return result
}

func genCode(templateContent string, relativePath string, outputFileName string, templateData any) {
	tmpl, err := template.New("ts").Parse(templateContent)
	if err != nil {
		panic(err.Error())
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		panic(err.Error())
	}
	code := buf.Bytes()

	fullPath := directory + relativePath
	if fullPath != "" && !strings.HasSuffix(fullPath, "/") {
		fullPath += "/"
	}

	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			panic(err.Error())
		}
	}
	err = os.WriteFile(fullPath+outputFileName+".ts", code, 0644)
	if err != nil {
		panic(err.Error())
	}
}

func firstLetterToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	// 将字符串转换为rune切片以正确处理Unicode字符
	runes := []rune(s)
	// 将第一个rune转为小写
	runes[0] = unicode.ToLower(runes[0])

	// 转换回字符串并返回
	return string(runes)
}

func firstLetterToUpper(s string) string {
	if len(s) == 0 {
		return s
	}
	// 将字符串转换为rune切片以正确处理Unicode字符
	runes := []rune(s)
	// 将第一个rune转为大写
	runes[0] = unicode.ToUpper(runes[0])

	// 转换回字符串并返回
	return string(runes)
}
