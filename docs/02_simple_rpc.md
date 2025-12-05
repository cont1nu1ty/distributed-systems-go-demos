# 02 - Simple RPC 框架实现

## 概述

本示例实现了一个最小化的 RPC（Remote Procedure Call）框架，演示了 RPC 的核心工作原理：如何将本地接口调用转换为远程通信，并在服务端还原为本地方法调用。

## RPC 核心概念

### Stub 与 Skeleton

```
┌──────────────────────────────────────────────────────────────────┐
│                         RPC 调用流程                              │
└──────────────────────────────────────────────────────────────────┘

Client Side                    Network                Server Side
┌─────────────┐                                      ┌─────────────┐
│   Client    │                                      │   Server    │
│             │                                      │             │
│  calc.Add() │                                      │             │
│      ↓      │                                      │             │
│  ┌────────┐ │                                      │  ┌────────┐ │
│  │  Stub  │ │  ──── JSON Request over TCP ────▶   │  │Skeleton│ │
│  │(Client)│ │                                      │  │(Server)│ │
│  └────────┘ │                                      │  └────────┘ │
│      ↓      │                                      │      ↓      │
│  Serialize  │                                      │  Deserialize│
│      ↓      │                                      │      ↓      │
│  Send TCP   │                                      │  Reflect    │
│      ↓      │                                      │      ↓      │
│  Wait Resp  │                                      │  Call Local │
│      ↓      │                                      │    Method   │
│  ┌────────┐ │  ◀──── JSON Response over TCP ────  │  ┌────────┐ │
│  │Deserial│ │                                      │  │Serialize│ │
│  └────────┘ │                                      │  └────────┘ │
│      ↓      │                                      │             │
│   return    │                                      │             │
└─────────────┘                                      └─────────────┘
```

## 架构设计

### 三层架构

```
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│  (CalculatorService interface & implementation)          │
└───────────────────┬─────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────┐
│                      RPC Layer                           │
│  ┌──────────────┐              ┌──────────────┐         │
│  │    Client    │              │    Server    │         │
│  │    (Stub)    │              │  (Skeleton)  │         │
│  │              │              │              │         │
│  │ - Call()     │              │ - Register() │         │
│  │ - Encode     │              │ - Invoke()   │         │
│  │ - Decode     │              │ - Reflect    │         │
│  └──────────────┘              └──────────────┘         │
└───────────────────┬─────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────┐
│                   Transport Layer                        │
│              (TCP + JSON Codec)                          │
└─────────────────────────────────────────────────────────┘
```

## 核心代码解析

### 1. 接口定义（契约）

```go
// 接口即契约 - 定义了远程服务的能力
type CalculatorService interface {
    Add(a, b int) int
    Multiply(a, b int) int
    Subtract(a, b int) int
    Divide(a, b int) int
}
```

**关键点**:
- 接口定义了方法签名，客户端和服务端都遵守
- 方法名和参数类型必须完全一致
- 这是 RPC 的"远程契约"

### 2. 客户端（Stub）- 隐藏网络细节

```go
// Call 方法隐藏了所有网络通信细节
func (c *Client) Call(service, method string, params ...interface{}) (interface{}, error) {
    // 1. 生成请求 ID（用于匹配响应）
    reqID := fmt.Sprintf("%s-%s-%d", service, method, time.Now().UnixNano())
    
    // 2. 创建响应通道
    respChan := make(chan *Response, 1)
    c.pending[reqID] = respChan
    
    // 3. 构造请求
    req := Request{
        ID:      reqID,
        Service: service,
        Method:  method,
        Params:  params,
    }
    
    // 4. 序列化并发送
    reqData, _ := EncodeRequest(req)
    c.conn.Write(reqData)
    
    // 5. 等待响应（阻塞）
    resp := <-respChan
    return resp.Result, nil
}
```

**关键点**:
- 客户端调用 `Call()` 就像调用本地方法一样
- 内部自动处理：序列化、发送、等待、反序列化
- 使用 channel 实现异步响应的同步等待

### 3. 服务端（Skeleton）- 反射调用

```go
// invoke 使用反射动态调用注册的服务方法
func (s *Server) invoke(serviceName, methodName string, params []interface{}) (interface{}, error) {
    // 1. 查找注册的服务实例
    service := s.services[serviceName]
    
    // 2. 获取服务的反射值
    serviceValue := reflect.ValueOf(service)
    
    // 3. 通过方法名获取方法
    method := serviceValue.MethodByName(methodName)
    
    // 4. 转换参数为反射值
    args := make([]reflect.Value, len(params))
    for i, param := range params {
        // JSON 数字默认是 float64，需要转换为 int
        if v, ok := param.(float64); ok {
            args[i] = reflect.ValueOf(int(v))
        }
    }
    
    // 5. 调用方法
    results := method.Call(args)
    
    // 6. 返回结果
    return results[0].Interface(), nil
}
```

**关键点**:
- `reflect.ValueOf()` 获取服务实例的反射对象
- `MethodByName()` 通过字符串查找方法
- `Call()` 动态调用方法
- 这就是 RPC "远程过程调用" 的本质

### 4. 并发响应匹配

```go
// 客户端维护一个 pending map 用于匹配请求和响应
type Client struct {
    pending map[string]chan *Response  // key: requestID
}

// 响应处理 goroutine
func (c *Client) handleResponses() {
    for {
        // 读取响应
        resp := decodeResponse()
        
        // 根据 ID 找到对应的请求通道
        respChan := c.pending[resp.ID]
        
        // 发送响应（唤醒等待的 Call）
        respChan <- resp
    }
}
```

**关键点**:
- 通过请求 ID 关联请求和响应
- 使用 channel 实现同步等待
- 支持并发请求（多个 goroutine 同时调用）

## 运行步骤

### 1. 启动 RPC 服务器
```bash
cd cmd/02_simple_rpc/server
go run main.go
```

**预期输出:**
```
2024/12/05 07:10:00 Registered service: CalculatorService
2024/12/05 07:10:00 RPC Server listening on :9100
```

### 2. 运行 RPC 客户端
```bash
cd cmd/02_simple_rpc/client
go run main.go
```

**预期输出:**
```
Simple RPC Client Demo
======================

--- Single Request Demo ---
Add(5, 3) = 8

--- Concurrent Requests Demo ---
[Request 1] Add(10, 20) = 30
[Request 2] Multiply(4, 5) = 20
[Request 3] Subtract(100, 50) = 50
[Request 4] Add(7, 8) = 15
[Request 5] Multiply(3, 3) = 9
[Request 6] Divide(100, 4) = 25

--- All requests completed ---
```

## RPC 对比 Raw Socket

| 特性 | Raw Socket (01) | Simple RPC (02) |
|------|----------------|-----------------|
| **调用方式** | 手动构造 JSON 请求 | `client.Call("Service", "Method", args)` |
| **类型安全** | 弱（字符串和 interface{}） | 中（接口定义） |
| **路由** | 手动 switch-case | 反射自动路由 |
| **序列化** | 手动实现 | 框架封装 |
| **并发支持** | 需要手动管理连接 | 内置并发响应匹配 |
| **代码量** | 服务端需要大量路由代码 | 只需实现接口 |

## RPC 框架的核心价值

### 1. 隐藏网络复杂性
- ✓ 调用远程服务就像调用本地方法
- ✓ 自动处理序列化、网络传输、反序列化
- ✓ 透明的错误传播

### 2. 接口即契约
- ✓ 接口定义了服务能力
- ✓ 类型系统提供编译时检查（在 gRPC 中更强）
- ✓ 服务端和客户端共享同一个接口定义

### 3. 反射实现动态分发
- ✓ 服务端无需手写路由代码
- ✓ 通过 `reflect` 动态调用方法
- ✓ 新增方法只需在接口中定义

### 4. 支持并发
- ✓ 一个连接可以并发多个请求
- ✓ 通过请求 ID 匹配响应
- ✓ 使用 channel 实现同步等待

## 局限性

虽然我们实现了一个基础的 RPC 框架，但与工业级框架相比，它还缺少：

- ✗ **服务发现**: 客户端仍需硬编码服务器地址
- ✗ **负载均衡**: 无法自动分发到多个服务实例
- ✗ **超时控制**: 虽然定义了 timeout，但未实现
- ✗ **重试机制**: 失败无法自动重试
- ✗ **协议优化**: JSON 性能较低，应使用 Protobuf
- ✗ **流式传输**: 只支持请求-响应，不支持流
- ✗ **认证授权**: 无安全机制
- ✗ **监控追踪**: 无法追踪请求链路

## 下一步

- 查看 [03_message_broker.md](./03_message_broker.md) 了解消息中间件模式
- 查看 [04_real_world_examples.md](./04_real_world_examples.md) 了解 gRPC 等工业级 RPC 框架
