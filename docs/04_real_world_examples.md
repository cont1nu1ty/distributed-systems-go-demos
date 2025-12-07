# 04 - 工业级示例：gRPC 与 NATS

## 概述

本部分展示工业级的 RPC 框架（gRPC）和消息系统（NATS），并与我们自实现的版本进行对比，帮助理解**教学级实现**与**生产级系统**的差距。

---

## Part 1: gRPC 示例

### 什么是 gRPC？

gRPC 是 Google 开源的高性能 RPC 框架，特点：
- ✓ 基于 **HTTP/2** 协议
- ✓ 使用 **Protocol Buffers** 序列化
- ✓ 支持**流式传输**（单向流、双向流）
- ✓ **多语言支持**（Go, Java, Python, C++, etc.）
- ✓ **强类型**（编译时检查）

### gRPC vs Simple RPC

| 特性 | Simple RPC (02) | gRPC |
|------|----------------|------|
| **协议** | TCP + JSON | HTTP/2 + Protobuf |
| **性能** | 中等 | 极高 |
| **类型安全** | 运行时 | 编译时 |
| **代码生成** | 无 | 自动生成 |
| **流式传输** | 不支持 | 支持 |
| **多语言** | 手动实现 | 自动支持 |
| **社区** | 无 | Google + 庞大社区 |

### Protocol Buffers 优势

#### 1. 强类型定义

```protobuf
// calculator.proto
syntax = "proto3";

service CalculatorService {
  rpc Add(BinaryOperation) returns (Result);
  rpc Multiply(BinaryOperation) returns (Result);
}

message BinaryOperation {
  int32 a = 1;
  int32 b = 2;
}

message Result {
  int32 value = 1;
}
```

**优势**:
- ✓ 类型明确（int32, string, bool, etc.）
- ✓ 编译时检查，避免运行时错误
- ✓ 自动生成代码，避免手写序列化

#### 2. 高效序列化

| 格式 | 大小 | 速度 |
|------|------|------|
| JSON | 100% (基准) | 100% (基准) |
| Protobuf | ~30% | ~5x 快 |

**原理**:
- JSON: `{"a": 5, "b": 3}` → 需要解析字段名
- Protobuf: 二进制编码 → 只有字段编号和值

#### 3. 向后兼容

```protobuf
// 版本 1
message BinaryOperation {
  int32 a = 1;
  int32 b = 2;
}

// 版本 2（添加新字段）
message BinaryOperation {
  int32 a = 1;
  int32 b = 2;
  string operation_id = 3;  // 新增字段
}
```

**好处**:
- ✓ 旧客户端可以忽略新字段
- ✓ 新客户端可以处理旧消息
- ✓ 无需强制升级所有服务

### HTTP/2 优势

#### 1. 多路复用

```
传统 HTTP/1.1:
┌─────────┐       ┌─────────┐       ┌─────────┐
│ Request1│──────▶│ Request2│──────▶│ Request3│
└─────────┘       └─────────┘       └─────────┘
   (阻塞)           (阻塞)           (阻塞)

HTTP/2:
┌─────────┐
│ Request1│─┐
└─────────┘ │
┌─────────┐ ├──▶ 同一个 TCP 连接
│ Request2│─┤    (并发传输)
└─────────┘ │
┌─────────┐ │
│ Request3│─┘
└─────────┘
```

#### 2. 流控制

- ✓ 自动调节发送速度
- ✓ 避免接收端过载
- ✓ 更高效的网络利用

### 代码生成流程

```bash
# 1. 定义 .proto 文件
# api/proto/calculator.proto

# 2. 生成 Go 代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/calculator.proto

# 3. 生成的文件
# - calculator.pb.go       (消息定义)
# - calculator_grpc.pb.go  (服务定义)
```

### 运行步骤

#### 1. 生成代码（已完成）

项目已包含生成的代码，如需重新生成：
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/calculator.proto
```

#### 2. 启动 gRPC 服务器

```bash
cd cmd/04_real_world_examples/grpc_example/server
go run main.go
```

**预期输出:**
```
2024/12/05 07:20:00 gRPC Server starting...
2024/12/05 07:20:00 gRPC Server listening on :50051
2024/12/05 07:20:00 Using HTTP/2 and Protocol Buffers
```

#### 3. 运行 gRPC 客户端

```bash
cd cmd/04_real_world_examples/grpc_example/client
go run main.go
```

**预期输出:**
```
gRPC Client Demo
================

--- Single Request Demo ---
Add(5, 3) = 8

--- Concurrent Requests Demo ---
[Request 1] Add(10, 20) = 30
[Request 2] Multiply(4, 5) = 20
[Request 3] Subtract(100, 50) = 50
...
```

---

## Part 2: NATS 示例

### 什么是 NATS？

NATS 是一个高性能的云原生消息系统，特点：
- ✓ **极高性能**（百万级 msg/s）
- ✓ **轻量级**（~20MB 内存占用）
- ✓ **原生集群**（自动路由和负载均衡）
- ✓ **多种模式**（Pub/Sub, Request/Reply, Queue Groups）
- ✓ **JetStream**（持久化支持）

### NATS vs 自实现 Broker

| 特性 | Broker (03) | NATS |
|------|-------------|------|
| **部署** | 内存（单进程） | 独立服务器 |
| **性能** | 极高（内存） | 百万级 msg/s |
| **持久化** | 无 | JetStream 支持 |
| **集群** | 无 | 原生集群 |
| **高可用** | 无 | 内置 HA |
| **监控** | 无 | 丰富的监控指标 |
| **客户端** | 自实现 | 多语言官方库 |
| **适用** | 教学 | 生产环境 |

### NATS 核心特性

#### 1. Subject-Based Messaging

```
news.tech.golang      → 接收：news.*, news.tech.*, news.tech.golang
updates.v2.release    → 接收：updates.*, updates.v2.*, updates.v2.release
alerts.critical.db    → 接收：alerts.*, alerts.critical.*, alerts.critical.db
```

**通配符**:
- `*` : 匹配单级（`news.*` 匹配 `news.tech`）
- `>` : 匹配多级（`news.>` 匹配 `news.tech.golang`）

#### 2. Queue Groups（负载均衡）

```
┌──────────┐
│Publisher │
└────┬─────┘
     │
     ▼
┌─────────────────┐
│  NATS Server    │
│  Topic: "work"  │
└────┬──────┬─────┘
     │      │
     ▼      ▼
┌─────────┐ ┌─────────┐
│Worker 1 │ │Worker 2 │  ← Queue Group "workers"
└─────────┘ └─────────┘
  (msg 1)     (msg 2)    只有一个 worker 收到每条消息
```

#### 3. Request-Reply 模式

```go
// Server (Responder)
nc.Subscribe("help", func(msg *nats.Msg) {
    response := "I can help!"
    msg.Respond([]byte(response))
})

// Client (Requester)
reply, _ := nc.Request("help", []byte("Need help!"), 1*time.Second)
```

### 安装 NATS Server

#### 方法 1: Docker（推荐）

```bash
docker run -p 4222:4222 -p 8222:8222 nats:latest
```

#### 方法 2: 下载二进制

```bash
# macOS
brew install nats-server

# Linux
wget https://github.com/nats-io/nats-server/releases/download/v2.10.7/nats-server-v2.10.7-linux-amd64.tar.gz
tar -xzvf nats-server-v2.10.7-linux-amd64.tar.gz
./nats-server-v2.10.7-linux-amd64/nats-server
```

### 运行步骤

#### 1. 启动 NATS Server

```bash
docker run -p 4222:4222 nats
```

**预期输出:**
```
[1] 2024/12/05 07:25:00.000000 [INF] Starting nats-server
[1] 2024/12/05 07:25:00.000000 [INF] Version:  2.10.7
[1] 2024/12/05 07:25:00.000000 [INF] Server is ready
[1] 2024/12/05 07:25:00.000000 [INF] Listening for client connections on 0.0.0.0:4222
```

#### 2. 启动订阅者（可以启动多个）

```bash
# 终端 2: 订阅 "news"
cd cmd/04_real_world_examples/nats_example/subscriber
go run main.go news 1

# 终端 3: 订阅 "news"（演示广播）
go run main.go news 2

# 终端 4: 订阅 "updates"
go run main.go updates 3
```

**预期输出:**
```
NATS Subscriber Demo
====================
Subscriber 1 starting...
Subscribing to subject: 'news'
Connecting to NATS server at nats://localhost:4222
Connected to NATS server
[Subscriber 1] Subscribed to subject 'news'
Waiting for messages... (Press Ctrl+C to exit)
```

#### 3. 运行发布者

```bash
cd cmd/04_real_world_examples/nats_example/publisher
go run main.go
```

**预期输出:**
```
NATS Publisher Demo
===================

--- Publishing Messages ---
[Message 1] Publishing to 'news': Breaking: Go 1.22 released!
[Message 1] Published successfully
...
```

**订阅者同时收到:**
```
[Subscriber 1] Received message #1 on subject 'news': Breaking: Go 1.22 released!
[Subscriber 2] Received message #1 on subject 'news': Breaking: Go 1.22 released!
```

### NATS 高级特性

#### JetStream（持久化）

```go
// 创建 Stream
js, _ := nc.JetStream()
js.AddStream(&nats.StreamConfig{
    Name:     "ORDERS",
    Subjects: []string{"orders.*"},
})

// 持久化发布
js.Publish("orders.new", []byte("order-123"))

// 持久化订阅（即使消费者离线也不丢消息）
js.Subscribe("orders.*", func(msg *nats.Msg) {
    msg.Ack()  // 确认消息
})
```

#### 集群模式

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  NATS-1     │◀──▶│  NATS-2     │◀──▶│  NATS-3     │
│  :4222      │    │  :4222      │    │  :4222      │
└─────────────┘    └─────────────┘    └─────────────┘
      ▲                  ▲                  ▲
      │                  │                  │
  ┌───┴───┐          ┌───┴───┐          ┌───┴───┐
  │Client1│          │Client2│          │Client3│
  └───────┘          └───────┘          └───────┘

- 自动故障转移
- 负载均衡
- 数据复制
```

---

## 总结

### RPC：gRPC vs Simple RPC

**何时使用 gRPC**:
- ✓ 微服务架构（服务间通信）
- ✓ 需要高性能（低延迟、高吞吐）
- ✓ 多语言环境
- ✓ 需要流式传输
- ✓ 强类型要求

**Simple RPC 的价值**:
- ✓ 理解 RPC 原理
- ✓ 教学演示
- ✓ 快速原型

### 消息系统：NATS vs Broker

**何时使用 NATS**:
- ✓ 分布式系统（跨进程通信）
- ✓ 需要持久化（JetStream）
- ✓ 需要高可用（集群）
- ✓ 复杂路由（通配符、Queue Groups）
- ✓ 生产环境

**Broker 的价值**:
- ✓ 理解 Pub/Sub 原理
- ✓ 单进程内通信
- ✓ 教学演示

### 技术选型建议

```
┌─────────────────────────────────────────────────────────┐
│              分布式通信技术选型                          │
└─────────────────────────────────────────────────────────┘

同步请求-响应:
  - 低延迟要求    → gRPC
  - 简单场景      → REST API
  - 学习理解      → Simple RPC

异步事件驱动:
  - 高吞吐量      → NATS / Kafka
  - 复杂路由      → NATS
  - 持久化存储    → Kafka / RabbitMQ
  - 学习理解      → Broker (Channel)

混合场景:
  - 微服务架构    → gRPC (服务间) + NATS (事件)
  - 实时系统      → gRPC (API) + NATS (推送)
```

---

## 下一步学习

1. **深入 gRPC**:
   - 流式传输（Server Streaming, Client Streaming, Bidirectional）
   - 拦截器（Interceptors）
   - 负载均衡（gRPC Load Balancing）

2. **深入 NATS**:
   - JetStream 高级特性
   - NATS Clustering
   - Monitoring & Observability

3. **生产实践**:
   - 服务网格（Service Mesh）
   - 可观测性（Tracing, Metrics, Logging）
   - 容错模式（Retry, Circuit Breaker, Timeout）
