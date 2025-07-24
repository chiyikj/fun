# Fun Websocket Framework

Fun 是一个基于 WebSocket 的实时通信框架，旨在简化前后端交互和微服务架构的开发。

官方网站: https://fun.cyi.cc/

## 🌟 特性

- **WebSocket 通信**: 基于 WebSocket 实现高性能实时通信
- **依赖注入**: 自动化依赖注入，简化组件间依赖管理
- **类型安全**: 强类型检查，提供编译时安全保障
- **代码生成**: 自动生成 TypeScript 客户端代码
- **拦截器支持**: 支持 Guard 拦截器，实现权限验证等通用逻辑
- **结构化数据传输**: 支持复杂数据结构的序列化和反序列化

## 📦 安装

```bash
go get github.com/chiyikj/fun
```


## 🚀 快速开始

### 1. 定义服务

```go
// userService.go
package userService

import (
    "fun"
)

type UserService struct {
    fun.Ctx
    // 其他依赖字段
}

type User struct {
    User string
    Name *string
}

func (ctx UserService) HelloWord(user User) *int8 {
    // 业务逻辑
    return nil
}

func init() {
    fun.BindService(UserService{})
}
```


### 2. 启动服务

```go
// main.go
package main

import (
    "fun"
    _ "your-module/service/userService" // 导入你的服务
)

func main() {
    fun.Gen()      // 生成客户端代码
    fun.Start(3000) // 启动服务在端口 3000
}
```


## 🛠 核心概念

### 服务 (Service)

服务是业务逻辑的载体，每个服务结构体必须嵌入 `fun.Ctx` 作为第一个字段：

```go
type UserService struct {
    fun.Ctx
    // 其他依赖字段
}
```


### 上下文 (Ctx)

`fun.Ctx` 提供了请求上下文信息：

- `Ip`: 客户端 IP 地址
- `Id`: 客户端唯一标识
- `State`: 状态信息
- `RequestId`: 请求唯一标识
- `Send`: 发送数据给客户端的方法
- `Close`: 关闭请求连接的方法

### 依赖注入

通过 `fun:"auto"` 标签实现自动依赖注入：

```go
type Config struct {
    Page int8
}

type X struct {
    Name   string
    Config *Config `fun:"auto"`
}

func (config *Config) New() {
    config.Page = 5
}
```


### 拦截器 (Guard)

拦截器用于在方法执行前进行验证或预处理：

```go
type AuthGuard struct {
    Config *Config `fun:"auto"`
}

func (g AuthGuard) Guard(serviceName string, methodName string, state map[string]string) *fun.Result[any] {
    // 实现权限验证逻辑
    return nil // 返回 nil 表示验证通过
}

// 绑定服务时添加拦截器
func init() {
    fun.BindService(UserService{}, AuthGuard{})
}
```


### 代码生成

框架支持自动生成 TypeScript 客户端代码：

```go
func main() {
    fun.Gen() // 自动生成前端 TypeScript 代码到 dist 目录
}
```


## 📞 API 使用

### 启动服务

```go
// 启动 HTTP WebSocket 服务
fun.Start(3000)

// 启动 HTTPS WebSocket 服务
fun.StartTls("cert.pem", "key.pem", 3000)
```


### 绑定服务

```go
// 绑定服务和全局拦截器
fun.BindService(UserService{}, AuthGuard{})
```


### 客户端调用

生成的 TypeScript 客户端可以这样使用：

```typescript
import fun from "./dist/fun";

const api = fun.create("ws://localhost:3000");
const result = await api.UserService.HelloWord({User: "test"});
```


## 🧪 测试

框架提供了便捷的测试工具：

```go
type UserServiceTest struct {
    Service UserService
}

func (test UserServiceTest) HelloWord() {
    request := fun.GetRequestInfo(map[string]any{
        "User": "test",
        "Name": "1212",
    }, map[string]string{})
    
    result := fun.MockRequest[*int8](
        fun.GetClientInfo("123456"), 
        request
    )
    
    // 验证结果
    fmt.Println(result)
}

func main() {
    fun.Test(UserServiceTest{})
}
```


## 📁 项目结构

```
your-project/
├── dist/                 # 自动生成的 TypeScript 客户端代码
├── service/
│   └── userService/
│       └── userService.go # 服务定义
├── main.go              # 主程序入口
└── go.mod
```


## 📘 详细文档

### 方法类型

Fun 框架支持两种方法类型：

1. **普通方法**: 返回结果给客户端
```go
func (ctx UserService) GetData() *UserData {
    // 返回数据
    return &UserData{...}
}
```


2. **代理方法**: 支持推送更新
```go
func (ctx UserService) WatchData(proxy fun.ProxyClose) *UserData {
    // 当需要推送更新时调用 ctx.Send()
    go func() {
        // 模拟异步推送
        time.Sleep(time.Second)
        ctx.Send(ctx.Id, ctx.RequestId, &UserData{...})
    }()
    
    // 返回初始数据
    return &UserData{...}
}
```


### 错误处理

```go
func (ctx UserService) GetData() *UserData {
    // 返回自定义错误
    if someCondition {
        return fun.Error(404, "User not found")
    }
    
    // 返回数据
    return &UserData{...}
}
```


## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 📄 许可证

本项目采用 MIT 许可证，详情请见 [LICENSE](LICENSE) 文件。