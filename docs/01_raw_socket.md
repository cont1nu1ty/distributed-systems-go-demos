# 01 - Raw Socket 通信示例

## 概述

本示例演示了使用 Go 标准库 `net` 包进行原生 TCP Socket 通信的实现。通过这个示例，您将了解到为什么直接使用 Socket 不适合构建复杂的分布式系统。

## 架构图

```
┌─────────────┐                           ┌─────────────┐
│   Client 1  │                           │             │
│             │──── TCP Connection ───────▶             │
└─────────────┘                           │             │
                                          │   Server    │
┌─────────────┐                           │   :9001     │
│   Client 2  │                           │             │
│             │──── TCP Connection ───────▶ Goroutine   │
└─────────────┘                           │  per conn   │
                                          │             │
┌─────────────┐                           │             │
│   Client N  │                           │             │
│             │──── TCP Connection ───────▶             │
└─────────────┘                           └─────────────┘
```

## 请求/响应协议

### Request (JSON)
```json
{
  "id": "req-1-1733384782123456",
  "service": "CalculatorService",
  "method": "Add",
  "params": [5, 3],
  "timeout_ms": 2000
}
```

### Response (JSON)
```json
{
  "id": "req-1-1733384782123456",
  "result": 8,
  "error": ""
}
```

## 运行步骤

### 1. 启动服务器
```bash
cd cmd/01_raw_socket/server
go run main.go
```

**预期输出:**
```
2024/12/05 07:00:00 Raw Socket Server listening on :9001
2024/12/05 07:00:00 Waiting for connections...
```

### 2. 运行客户端
在新终端中:
```bash
cd cmd/01_raw_socket/client
go run main.go
```

**预期输出:**
```
2024/12/05 07:00:05 Raw Socket Client Demo
2024/12/05 07:00:05 ======================

--- Single Request Demo ---
2024/12/05 07:00:05 [Client 1] Sending request: CalculatorService.Add([5 3])
2024/12/05 07:00:05 [Client 1] Response: 8

--- Concurrent Requests Demo ---
2024/12/05 07:00:05 [Client 1] Sending request: CalculatorService.Add([10 20])
2024/12/05 07:00:05 [Client 2] Sending request: CalculatorService.Multiply([4 5])
2024/12/05 07:00:05 [Client 3] Sending request: CalculatorService.Subtract([100 50])
...
```

## 核心代码解析

### 服务器端 - 手动处理每个连接

```go
// 每个连接需要一个独立的 goroutine
for {
    conn, err := listener.Accept()
    if err != nil {
        log.Printf("Accept error: %v", err)
        continue
    }
    
    // 并发处理
    go handleConnection(conn, currentID)
}
```

### 协议处理 - 手动编解码

```go
// 读取 JSON 消息 - 需要处理粘包/拆包
func ReadJSON(conn net.Conn, v interface{}) error {
    reader := bufio.NewReader(conn)
    line, err := reader.ReadBytes('\n')  // 使用换行符分隔
    if err != nil {
        return err
    }
    
    return json.Unmarshal(line, v)
}
```

### 业务逻辑 - 手动路由

```go
// 必须手动解析服务名和方法名
if req.Service != "CalculatorService" {
    resp.Error = "unknown service"
    return resp
}

switch req.Method {
case "Add":
    resp.Result = req.Params[0] + req.Params[1]
case "Multiply":
    resp.Result = req.Params[0] * req.Params[1]
default:
    resp.Error = "unknown method"
}
```

## 痛点分析

### 1. 手动协议设计与实现
- ✗ 需要自己定义请求/响应格式
- ✗ 需要处理消息边界（粘包/拆包）
- ✗ 需要自己实现序列化/反序列化
- ✗ 协议变更需要客户端和服务端同步更新

### 2. 缺乏类型安全
- ✗ 参数是 `[]int`，无法强制类型检查
- ✗ 方法名是字符串，拼写错误在运行时才能发现
- ✗ 服务名硬编码，容易出错

### 3. 错误处理复杂
- ✗ 网络错误、序列化错误、业务错误混在一起
- ✗ 每一步都需要 `if err != nil`
- ✗ 错误信息传递需要自己设计

### 4. 缺少关键特性
- ✗ **无服务发现**: 客户端必须硬编码服务器地址
- ✗ **无负载均衡**: 无法自动分发请求到多个服务器
- ✗ **无超时控制**: 虽然定义了 `timeout_ms`，但需要自己实现
- ✗ **无重试机制**: 失败了就失败了
- ✗ **无熔断降级**: 服务不可用时没有保护机制
- ✗ **无监控追踪**: 请求链路无法追踪

### 5. 并发模型简陋
- ✗ 每个连接一个 goroutine，资源占用大
- ✗ 没有连接池
- ✗ 没有请求队列和流控

## 为什么不适合构建分布式系统？

| 问题 | Socket 需要自己解决 | RPC 框架已解决 |
|------|---------------------|----------------|
| 协议设计 | ✗ 手动设计 | ✓ 自动生成 |
| 序列化 | ✗ 手动实现 | ✓ 内置支持 |
| 服务注册 | ✗ 无 | ✓ 内置或集成 |
| 负载均衡 | ✗ 无 | ✓ 内置支持 |
| 超时控制 | ✗ 手动实现 | ✓ 内置支持 |
| 熔断降级 | ✗ 无 | ✓ 可集成 |
| 链路追踪 | ✗ 无 | ✓ 可集成 |
| 类型安全 | ✗ 弱 | ✓ 强 |

## 结论

原生 Socket 是**通信原语**，不是**分布式架构工具**。它适合：
- 学习网络编程基础
- 实现特定的低级协议
- 性能极致优化的场景（但需要大量工作）

对于构建分布式系统，应该使用：
- **RPC 框架**（如 gRPC、Thrift）：提供标准化的通信模式
- **消息中间件**（如 NATS、Kafka）：提供异步解耦能力

下一步请查看 [02_simple_rpc.md](./02_simple_rpc.md) 了解 RPC 如何解决这些问题。
