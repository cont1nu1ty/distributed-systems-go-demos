# 03 - Message Broker (消息代理) 实现

## 概述

本示例使用 Go 的 `channel` 实现了一个内存版的 Pub/Sub 消息代理系统，展示了消息中间件的核心工作原理，以及它与传统消息队列的本质区别。

## 核心概念：Pub/Sub vs Queue

### Pub/Sub 模型（发布/订阅）

```
┌─────────────┐                      ┌──────────────┐
│ Publisher 1 │                      │ Subscriber 1 │
│             │──┐                 ┌─▶              │
└─────────────┘  │                 │  └──────────────┘
                 │  ┌──────────┐   │
┌─────────────┐  │  │          │   │  ┌──────────────┐
│ Publisher 2 │──┼─▶│  Broker  │───┼─▶│ Subscriber 2 │
│             │  │  │  Topic:  │   │  │              │
└─────────────┘  │  │  "news"  │   │  └──────────────┘
                 │  │          │   │
┌─────────────┐  │  └──────────┘   │  ┌──────────────┐
│ Publisher 3 │──┘                 └─▶│ Subscriber 3 │
│             │                        │              │
└─────────────┘                        └──────────────┘

特点：
- 一条消息可以被多个订阅者接收（广播）
- 订阅者之间相互独立
- 消息按主题（Topic）路由
```

### Queue 模型（点对点队列）

```
┌─────────────┐                      ┌──────────────┐
│ Producer 1  │                      │  Consumer 1  │
│             │──┐                 ┌─▶ (处理 msg1)  │
└─────────────┘  │                 │  └──────────────┘
                 │  ┌──────────┐   │
┌─────────────┐  │  │          │   │  ┌──────────────┐
│ Producer 2  │──┼─▶│  Queue   │───┼─▶│  Consumer 2  │
│             │  │  │ [msg1,   │   │  │ (处理 msg2)  │
└─────────────┘  │  │  msg2,   │   │  └──────────────┘
                 │  │  msg3]   │   │
┌─────────────┐  │  └──────────┘   │  ┌──────────────┐
│ Producer 3  │──┘                 └─▶│  Consumer 3  │
│             │                        │ (处理 msg3)  │
└─────────────┘                        └──────────────┘

特点：
- 一条消息只能被一个消费者接收（竞争）
- 消费者之间竞争消息
- 用于负载均衡
```

## 架构设计

### Broker 核心结构

```go
type Broker struct {
    mu          sync.RWMutex
    subscribers map[string][]chan Message  // Topic -> []Channel
}

type Message struct {
    Topic     string
    Payload   interface{}
    Timestamp time.Time
}
```

### 工作流程

```
┌──────────────────────────────────────────────────────────────┐
│                    Broker 内部机制                            │
└──────────────────────────────────────────────────────────────┘

1. Subscribe (订阅)
   Client ──Subscribe("news")──▶ Broker
                                    │
                                    ├─ Create channel
                                    ├─ Add to subscribers["news"]
                                    └─ Return channel ─▶ Client

2. Publish (发布)
   Client ──Publish("news", msg)──▶ Broker
                                       │
                                       ├─ Get subscribers["news"]
                                       │
                                       ├─ For each subscriber:
                                       │   └─ Send msg to channel (non-blocking)
                                       │
                                       └─ Return

3. Receive (接收)
   Client ──<-msgChan──▶ Receive message from channel
```

## 核心代码解析

### 1. 订阅：创建 Channel

```go
func (b *Broker) Subscribe(topic string) (<-chan Message, error) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    // 为订阅者创建一个缓冲 channel
    ch := make(chan Message, 100)
    
    // 添加到订阅者列表
    b.subscribers[topic] = append(b.subscribers[topic], ch)
    
    return ch, nil
}
```

**关键点**:
- ✓ 每个订阅者都有自己的 channel
- ✓ 缓冲 channel 避免发送阻塞
- ✓ 返回的是只读 channel（`<-chan`）

### 2. 发布：非阻塞发送

```go
func (b *Broker) Publish(topic string, payload interface{}) error {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    msg := Message{
        Topic:     topic,
        Payload:   payload,
        Timestamp: time.Now(),
    }
    
    subscribers := b.subscribers[topic]
    
    // 并发发送给所有订阅者
    for _, ch := range subscribers {
        select {
        case ch <- msg:
            // 发送成功
        default:
            // Channel 已满，丢弃消息（可以改为阻塞或报错）
            log.Printf("Warning: subscriber channel full, message dropped")
        }
    }
    
    return nil
}
```

**关键点**:
- ✓ 使用 `select` + `default` 实现非阻塞发送
- ✓ 如果订阅者处理慢，不会阻塞发布者
- ✓ 这是 **不轮询** 的关键

### 3. 接收：使用 Channel

```go
// 客户端代码
msgChan, _ := broker.Subscribe("news")

for msg := range msgChan {
    // 处理消息
    fmt.Printf("Received: %v\n", msg.Payload)
}
```

**关键点**:
- ✓ 直接从 channel 接收，**没有轮询**
- ✓ `for range` 会阻塞等待新消息
- ✓ Channel 关闭时循环自动退出

## 为什么不需要轮询？

### 传统轮询方式（错误示范）

```go
// ❌ 错误：轮询检查是否有新消息
for {
    if hasMessage() {
        msg := getMessage()
        process(msg)
    }
    time.Sleep(100 * time.Millisecond)  // 浪费 CPU
}
```

### Go Channel 方式（正确）

```go
// ✓ 正确：Channel 阻塞等待
for msg := range msgChan {
    process(msg)  // 有消息时自动唤醒
}
```

**原理**:
- Channel 是 Go 运行时的同步原语
- 从空 channel 读取会阻塞 goroutine
- 有消息时自动唤醒，无 CPU 消耗

## 运行步骤

### 1. 启动 Broker 服务器
```bash
cd cmd/03_message_broker/broker
go run main.go
```

**预期输出:**
```
2024/12/05 07:15:00 Message Broker Server starting...
2024/12/05 07:15:00 Message Broker Server listening on :9200
2024/12/05 07:15:00 Waiting for connections...
```

### 2. 启动消费者（可以启动多个）
```bash
# 终端 2: 订阅 "news"
cd cmd/03_message_broker/consumer
go run main.go news 1

# 终端 3: 订阅 "news"（第二个消费者）
go run main.go news 2

# 终端 4: 订阅 "updates"
go run main.go updates 3
```

**预期输出 (Consumer 1):**
```
Message Broker Consumer Demo
=============================
Consumer 1 starting...
Subscribing to topic: 'news'
[Consumer 1] Subscribed to topic 'news'
```

### 3. 运行生产者
```bash
cd cmd/03_message_broker/producer
go run main.go
```

**预期输出 (Producer):**
```
Message Broker Producer Demo
=============================

--- Publishing Messages ---
[Message 1] Publishing to topic 'news': map[content:Go 1.22 released! title:Breaking News]
[Message 1] Published successfully
...
```

**预期输出 (Consumer 1 & 2 同时收到):**
```
[Consumer 1] Received message #1 on topic 'news': map[content:Go 1.22 released! title:Breaking News]
[Consumer 2] Received message #1 on topic 'news': map[content:Go 1.22 released! title:Breaking News]
```

## Pub/Sub vs Redis List (队列)

| 特性 | Pub/Sub (本示例) | Redis List (LPUSH/RPOP) |
|------|-----------------|-------------------------|
| **消息分发** | 所有订阅者都收到（广播） | 只有一个消费者收到（竞争） |
| **适用场景** | 事件通知、日志收集 | 任务队列、负载均衡 |
| **消费者数量** | 多个独立消费者 | 多个竞争消费者 |
| **消息持久化** | 内存（本示例不持久化） | 可持久化 |
| **实现机制** | Channel 广播 | 队列出队 |

### 使用场景对比

**使用 Pub/Sub**:
- ✓ 系统事件通知（如用户注册事件）
- ✓ 实时日志收集（多个监控系统订阅）
- ✓ 聊天室消息（所有用户接收）
- ✓ 配置更新通知

**使用 Queue**:
- ✓ 任务分发（图片处理、邮件发送）
- ✓ 负载均衡（多个 worker 竞争任务）
- ✓ 顺序处理（FIFO）

## 与 NATS 的对比

| 特性 | 本示例 | NATS |
|------|--------|------|
| **实现** | 内存 + Go Channel | 独立服务器 + 网络协议 |
| **持久化** | ✗ 无（重启丢失） | ✓ JetStream 支持 |
| **集群** | ✗ 单机 | ✓ 原生集群支持 |
| **性能** | 极高（内存） | 非常高（百万级 msg/s） |
| **适用场景** | 单进程、教学 | 分布式系统、生产环境 |
| **类型** | 教学级 | 工业级 |

## 总结

### Broker 的核心价值

1. **解耦**：发布者和订阅者互不感知
2. **广播**：一条消息可以被多个消费者接收
3. **异步**：发布者不需要等待消费者处理完成
4. **扩展**：可以动态增加订阅者

### Go Channel 的优势

1. **零轮询**：阻塞等待，无 CPU 浪费
2. **并发安全**：内置同步机制
3. **类型安全**：编译时检查消息类型
4. **简洁优雅**：几行代码实现复杂逻辑

## 下一步

查看 [04_real_world_examples.md](./04_real_world_examples.md) 了解工业级消息系统 NATS 和 RPC 框架 gRPC。
