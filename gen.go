package fun

import (
	"bytes"
	"os"
	"reflect"
	"slices"
	"strings"
	"text/template"
)

var directory = "dist/"

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

type genImportType struct {
	Name string
	Path string
}

type genServiceType struct {
	ServiceName       string
	GenMethodTypeList []*genMethodType
	GenImport         []*genImportType
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

func genService(
	service *service,
	serviceContext *genServiceType,
	visitedStructPaths []string,
) {
	// 收集服务方法中涉及的结构体导入路径
	var nestedImports []*genImportType

	// 遍历服务中的每个方法
	for _, method := range service.methodList {
		var returnValueText string
		var dtoText string
		var argsText string
		var genericTypeText string

		// 处理返回值类型
		if method.method.Type.NumOut() == 0 {
			returnValueText = "null"
		} else {
			returnType := method.method.Type.Out(0)

			// 转换为 TypeScript 类型
			returnValueText = typeToJsType(returnType)

			// 如果是结构体类型，递归生成导入路径
			if returnType.Kind() == reflect.Ptr {
				returnType = returnType.Elem()
			}
			if returnType.Kind() == reflect.Struct {
				nestedImports = append(nestedImports, genStruct(returnType, visitedStructPaths))
			}
		}

		// 处理 DTO 参数
		if method.dto != nil {
			dtoText += "dto:" + typeToJsType(*method.dto)
			argsText += ",dto"
			nestedImports = append(nestedImports, genStruct(*method.dto, visitedStructPaths))
		} else {
			argsText += ",null"
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
			MethodName:      method.method.Name,
			ReturnValueText: returnValueText,
			DtoText:         dtoText,
			ArgsText:        argsText,
			GenericTypeText: genericTypeText,
		})
	}

	// 去重导入路径
	serviceContext.GenImport = deduplicateServiceImports(nestedImports)

	// 生成 TypeScript 文件
	genCode(
		genServiceTemplate(),
		"",
		service.serviceType.Name(),
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
	genContext := gen{GenServiceList: []*genServiceType{}}

	for _, service := range f.serviceList {
		serviceContext := &genServiceType{
			ServiceName:       service.serviceType.Name(),
			GenMethodTypeList: []*genMethodType{},
		}

		genContext.GenServiceList = append(genContext.GenServiceList, serviceContext)
		genService(service, serviceContext, visitedStructPaths)
	}

	genCode(genDefaultServiceTemplate(), "", "fun", genContext)
}

func genStruct(t reflect.Type, visitedPaths []string) *genImportType {
	// 提取结构体所在的包路径并生成相对路径
	pkgParts := strings.Split(t.PkgPath(), "/")
	relativePath := strings.Join(pkgParts[1:], "/")

	// 如果路径已生成过，直接返回引用
	if slices.Contains(visitedPaths, relativePath) {
		return &genImportType{Name: t.Name(), Path: relativePath}
	}

	// 创建结构体模板数据
	structTemplate := genClassType{
		Name: t.Name(),
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
		structTemplate.GenClassFieldType = append(structTemplate.GenClassFieldType, &genClassFieldType{
			Name: name,
			Type: jsType,
		})

		// 如果字段是结构体，递归生成导入路径
		if fieldType.Kind() == reflect.Struct {
			nestedImports = append(nestedImports, genStruct(fieldType, visitedPaths))
		}

		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Struct {
			nestedImports = append(nestedImports, genStruct(fieldType.Elem(), visitedPaths))
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
		t.Name(),
		structTemplate,
	)

	// 标记该路径已生成
	visitedPaths = append(visitedPaths, relativePath)

	// 返回结构体导入引用（含完整路径）
	return &genImportType{
		Name: t.Name(),
		Path: relativePath + "/" + t.Name(),
	}
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
			Name: imp.Name,
			Path: relativePathPrefix + strings.Join(impPathParts[commonPrefixLen:], "/"),
		})
	}

	return result
}

func genCode(templateContent string, relativePath string, outputFileName string, templateData any) {
	tmpl, err := template.New("ts").Parse(templateContent)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		panic(err)
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
			panic(err)
		}
	}

	err = os.WriteFile(fullPath+outputFileName+".ts", code, 0644)
	if err != nil {
		panic(err)
	}
}
