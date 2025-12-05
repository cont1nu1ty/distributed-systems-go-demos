# Role: 资深 Go 语言分布式系统架构师和技术布道师

# Context: 我需要创建一个教育性的 Go 项目，用于向他人（如团队成员、学生）清晰地解释分布式系统的几个核心概念。这个项目将作为我技术报告的核心实践依据，因此代码必须遵循 Go 的标准实践、体现语言特性，并且每个部分都紧密对应一个理论问题。

# Goal: 生成一个完整的、模块化的 Go 项目，该项目通过四个独立的、可运行的示例，逐一演示并回答以下四个问题：
1.  为什么传统的 Socket 通信不适合直接用于构建复杂的分布式系统？
2.  典型的 RPC（远程过程调用）中间件是如何工作的？请演示其将本地调用转化为远程通信，并在服务端还原为本地调用的完整过程。
3.  不同 RPC 框架使用的传输协议有何共性与差异？（通过代码注释和文档说明）
4.  如何实现一个消息中间件？并对比它与纯消息队列在使用上的差异。（以 NATS 作为工业级消息系统为例）

# Instructions: 请遵循 Go 的标准项目布局和最佳实践来创建项目。每个子目录都是一个独立的演示模块，包含其自身的代码和说明文档。

## 项目结构
distributed-systems-go-demos/
├── go.mod # Go 模块定义文件
├── README.md # (主报告) 综合性报告，回答所有四个问题，并链接到各个示例
├── cmd/ # 各个示例的可执行入口
│ ├── 01_raw_socket/
│ │ ├── server/
│ │ │ └── main.go
│ │ └── client/
│ │ └── main.go
│ ├── 02_simple_rpc/
│ │ ├── server/
│ │ │ └── main.go
│ │ └── client/
│ │ └── main.go
│ ├── 03_message_broker/
│ │ ├── broker/
│ │ │ └── main.go
│ │ ├── producer/
│ │ │ └── main.go
│ │ └── consumer/
│ │ └── main.go
│ └── 04_real_world_examples/
│ ├── grpc_example/
│ │ ├── server/
│ │ │ └── main.go
│ │ └── client/
│ │ └── main.go
│ └── nats_example/
│ ├── publisher/
│ │ └── main.go
│ └── subscriber/
│ └── main.go
├── internal/ # 内部共享库，不对外暴露
│ ├── rpc/ # 自实现的极简 RPC 框架
│ │ ├── client.go # RPC 客户端 Stub
│ │ ├── server.go # RPC 服务端 Skeleton
│ │ └── codec.go # JSON 编解码器
│ └── broker/ # 自实现的消息代理
│ └── broker.go # Broker 核心逻辑
├── pkg/ # 可对外暴露的公共库
│ └── socket/ # 原生 Socket 通信辅助函数
│ └── util.go
├── api/ # API 定义文件 (如 .proto)
│ └── proto/
│ └── calculator.proto
└── docs/ # 各模块的详细文档
├── 01_raw_socket.md
├── 02_simple_rpc.md
├── 03_message_broker.md
└── 04_real_world_examples.md



## 各部分详细指令

### 1. `cmd/01_raw_socket`
- **目标**: 直观展示原生 Socket 在 Go 中的实现和痛点。
- **实现**:
    - 使用 Go 标准库 `net` 包。
    - `server/main.go`: 监听 TCP 端口，使用 `goroutine` 为每个连接创建一个新的 handler。Handler 读取约定格式的数据（如 JSON `{"op": "add", "a": 5, "b": 3}`），处理后返回 JSON `{"result": 8}`。
    - `client/main.go`: 连接服务端，发送 JSON 请求，并解析响应。
- **关键点**: `docs/01_raw_socket.md` 中必须强调：手动协议处理、缺乏类型安全、复杂的错误处理（`if err != nil`）、没有服务发现和负载均衡等。

### 2. `cmd/02_simple_rpc` & `internal/rpc`
- **目标**: 揭示 RPC 的核心原理，并展示 Go 的接口和反射特性。
- **实现**:
    - `internal/rpc`:
        - 定义一个服务接口，如 `type CalculatorService interface { Add(a, b int) int }`。
        - `client.go`: 实现一个 RPC 客户端，它通过 `net.Conn` 调用远程服务。调用方法时，将方法名、参数序列化为 JSON 并发送。
        - `server.go`: 实现一个 RPC 服务端。它使用一个 map 来注册服务实例（`map[string]interface{}`）。接收到请求后，使用 `reflect` 包动态调用注册实例的对应方法。
    - `server/main.go`: 创建一个 `CalculatorService` 的具体实现，并将其注册到 RPC 框架中。
    - `client/main.go`: 像调用本地接口一样调用远程服务。
- **关键点**: `docs/02_simple_rpc.md` 中必须清晰标注：接口定义如何成为契约，`reflect.Value.MethodByName` 如何实现动态调用，以及这如何隐藏了底层网络通信的复杂性。

### 3. `cmd/03_message_broker` & `internal/broker`
- **目标**: 用 Go 的并发原语实现一个 Pub/Sub 模型的消息代理。
- **实现**:
    - `internal/broker/broker.go`:
        - 核心是一个 `Broker` 结构体，包含一个 `map[string][]chan Message`，用于存储每个主题的订阅者通道。
        - `Subscribe(topic string) <-chan Message`: 为订阅者创建一个新的 channel，并将其添加到对应主题的列表中。
        - `Publish(topic string, msg Message)`: 将消息发送到该主题的所有订阅者 channel。这个过程必须是并发的，不能阻塞。
    - `broker/main.go`: 启动 Broker 服务，可能通过 HTTP 或其他方式暴露 `Subscribe` 和 `Publish` 的能力。为简化，可以直接在内存中运行。
    - `producer/main.go` & `consumer/main.go`: 连接到 Broker，发布和订阅消息。
- **关键点**: `docs/03_message_broker.md` 中必须解释：`channel` 如何替代轮询，`select` 如何用于多路复用，以及这种 Pub/Sub 模型与 Redis 的 `LPUSH`/`RPOP`（点对点队列）的根本区别。

### 4. `cmd/04_real_world_examples`
- **目标**: 展示工业界标准，与自实现示例形成对比。
- **实现**:
    - `grpc_example`: 使用 `google.golang.org/grpc` 和 `protoc-gen-go`。展示 `.proto` 文件定义、生成的 Go 代码、服务端和客户端实现。在文档中强调 HTTP/2 和 Protobuf 的优势。
    - `nats_example`: 使用 `github.com/nats-io/nats.go`。演示发布/订阅模式，并与我们自实现的 Broker 对比，突出 NATS 的持久化、集群和高可用性。
- **关键点**: `docs/04_real_world_examples.md` 中需要总结：gRPC 适合同步的请求-响应模型，而 NATS/Kafka 等消息系统更适合异步的事件驱动架构。

# Constraints:
- **语言**: Go 1.21+
- **项目结构**: 严格遵循上述 Go 标准项目布局。
- **并发**: 必须使用 `goroutine` 和 `channel` 进行并发编程，严禁使用轮询。
- **接口**: 充分利用 Go 的 `interface` 来设计组件间的契约。
- **错误处理**: 严格遵循 Go 的 `if err != nil` 错误处理模式。
- **依赖管理**: 所有外部依赖必须在 `go.mod` 中声明。自实现的 RPC 和 Broker 只能使用 Go 标准库。
- **文档**: `docs/` 目录下的每个 `.md` 文件都应是对应模块的详细技术说明，可直接用于报告。

# Output Format:
1.  完整的项目文件树结构。
2.  `go.mod` 文件内容。
3.  每个源代码文件（`.go`）的完整内容。
4.  `api/proto/calculator.proto` 文件内容。
5.  `README.md` 和 `docs/` 目录下所有 Markdown 文件的详细内容
