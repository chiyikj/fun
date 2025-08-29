# Fun Framework

[![GoDoc](https://godoc.org/github.com/chiyikjk/fun?status.svg)](https://godoc.org/github.com/chiyikjk/fun) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

<a href="https://fun.chiyikjk.com">官方网站</a>

Fun 是一个基于 Go 语言的 WebSocket 框架，提供了服务绑定、依赖注入、参数验证、日志记录等功能。它专注于简化 WebSocket 应用开发，支持自动生成 TypeScript 客户端代码，让前后端开发更加高效。

## 特性

- 🚀 基于 Gorilla WebSocket 构建
- 🔧 自动依赖注入
- 🛡️ 参数验证（集成 validator.v10）
- 📝 多级别日志系统（支持文件和终端输出）
- 🔄 WebSocket 心跳检测和连接管理
- 🎯 自动生成 TypeScript 客户端代码
- 🧪 内置测试工具
- 🛡️ 守卫机制（类似中间件）
- 📦 结构化响应格式
- 🧭 枚举类型支持

## 快速开始

### 1. 安装

```bash
go get github.com/chiyikjk/fun
```

### 2. 创建简单服务

创建 `main.go`：

```go
package main

import (
    "github.com/chiyikjk/fun"
)

// 定义服务结构体
type HelloService struct {
    fun.Ctx // 必须嵌入 Ctx
}

// 定义 DTO
type HelloDto struct {
    Name string 
}

// 定义方法
func (s *HelloService) Hello(dto HelloDto) string {
    return "Hello, " + dto.Name
}

func main() {
    // 绑定服务
    fun.BindService(HelloService{})
    
    // 启动服务
    fun.Start(3000)
}
```

### 3. 运行服务

```bash
go run main.go
```

服务将在端口 3000 上启动。

### 4. 生成 TypeScript 客户端代码

```go
// 在 main.go 中添加
fun.Gen()
```

运行后，TypeScript 代码将生成在 `../gen/ts/` 目录下。

### 5. 客户端使用示例

```typescript
import fun from "./gen/ts/fun";

// 创建客户端实例
const client = fun.create("ws://localhost:3000");

// 调用服务方法
const result = await client.helloService.hello({ name: "World" });
console.log(result); // "Hello, World"
```

## 核心概念

### 服务（Service）

服务是包含业务逻辑的结构体，必须嵌入 [fun.Ctx](file://f:\fun\ctx.go#L2-L12)：

```go
type UserService struct {
    fun.Ctx // 提供上下文信息
    // 其他依赖...
}
```

### 方法（Method）

服务中的公开方法会自动暴露为 WebSocket 接口：

```go
func (s *UserService) GetUser(dto UserDto) *User {
    // 业务逻辑
    return &User{Name: dto.Name}
}
```

### DTO（Data Transfer Object）

用于传递参数的对象，支持参数验证：

```go
type UserDto struct {
    Name  string `validate:"required"`
    Email string `validate:"required,email"`
    Age   int    `validate:"min=0,max=150"`
}
```

### 枚举（Enum）

Fun 框架支持生成 TypeScript 枚举类型。要使用此功能，需要定义 `uint8` 类型并实现 `enum` 或 `displayEnum` 接口：

#### 基础枚举

```go
// 实现 enum 接口
type Status uint8

func (s Status) Names() []string {
    return []string{
        "Active",
        "Inactive",
    }
}
```

#### 显示枚举

```go
// 实现 displayEnum 接口
type UserStatus uint8

func (s UserStatus) Names() []string {
    return []string{
        "Active",
        "Inactive",
        "Pending",
    }
}

func (s UserStatus) DisplayNames() []string {
    return []string{
        "已激活",
        "未激活",
        "待审核",
    }
}
```

生成的 TypeScript 代码：

```typescript
enum userStatus {
  Active,
  Inactive,
  Pending,
}

export function userStatusDisplayName(value:userStatus): string {
  switch (value) {
    case userStatus.Active:
      return '已激活';
    case userStatus.Inactive:
      return '未激活';
    case userStatus.Pending:
      return '待审核';
    default:
      return "未知";
  }
}

export function userStatusDisplayNames(): string[] {
  return [
    '已激活',
    '未激活',
    '待审核',
  ];
}

export default userStatus
```

### 依赖注入

使用 `fun.Wired[T]()` 进行依赖注入：

```go
type Repository struct {
    // 数据库连接等
}

func (r *Repository) New() {
    // 初始化逻辑
}

type UserService struct {
    fun.Ctx
    Repo *Repository `fun:"auto"` // 自动注入
}
```

### 守卫（Guard）

类似中间件，用于请求前的处理：

```go
type AuthGuard struct {
    // 依赖...
}

func (g *AuthGuard) Guard(ctx fun.Ctx) {
    // 鉴权逻辑
    // 失败时可以 panic 错误
}

// 绑定全局守卫
fun.BindGuard(AuthGuard{})

// 或绑定服务级守卫
fun.BindService(UserService{}, AuthGuard{})
```

## 高级功能

### 代理模式（Proxy）

支持长连接模式，服务端可以主动推送数据：

```go
func (s *UserService) Subscribe(dto SubscribeDto, close fun.ProxyClose) *chan string {
    ch := make(chan string)
    
    // 处理关闭回调
    close(func() {
        close(ch)
    })
    
    go func() {
        for {
            select {
            case msg := <-ch:
                s.Push(s.Id, s.RequestId, msg) // 推送数据
            }
        }
    }()
    
    return &ch
}
```

### 日志系统

```go
// 配置日志
logger := &fun.Logger{
    Level:          fun.InfoLevel,
    Mode:           fun.FileMode,
    MaxSizeFile:    10,   // 10MB
    MaxNumberFiles: 100,  // 最多100个文件
    ExpireLogsDays: 7,    // 保留7天
}
fun.ConfigLogger(logger)

// 使用日志
fun.InfoLogger("服务启动成功")
fun.ErrorLogger("发生错误", err)
```

### 参数验证

集成 `validator.v10`，支持自定义验证规则：

```go
// 绑定自定义验证规则
fun.BindValidate("custom", func(fl validator.FieldLevel) bool {
    // 自定义验证逻辑
    return true
})

// 在 DTO 中使用
type CustomDto struct {
    Field string `validate:"custom"`
}
```

## 测试支持

提供完整的测试工具：

```go
func TestHelloService(t *testing.T) {
    // 创建请求
    request := fun.GetRequestInfo(t, HelloService{}, "Hello", HelloDto{Name: "World"}, nil)
    
    // 发起请求
    result := fun.MockRequest[string](t, request)
    
    // 验证结果
    if *result.Data != "Hello, World" {
        t.Errorf("期望 'Hello, World', 得到 '%s'", *result.Data)
    }
}
```

## 配置选项

### 服务配置

```go
// 启动 HTTP 服务
fun.Start(3000)

// 启动 HTTPS 服务
fun.StartTls("cert.pem", "key.pem", 3000)
```

### 日志配置

```go
type Logger struct {
    Level          uint8  // 日志级别
    Mode           uint8  // 输出模式 (TerminalMode/FileMode)
    MaxSizeFile    uint8  // 文件最大大小(MB)
    MaxNumberFiles uint64 // 文件最多数量
    ExpireLogsDays uint8  // 文件保留时间(天)
}
```
欢迎提交 Issue 和 Pull Request！