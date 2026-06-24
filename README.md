# httpPush

基于 HTTP 长连接的核心推送系统，支持集群部署。

## 特性

- **双协议接入**：同时支持 WebSocket 和 HTTP 长轮询两种客户端连接方式
- **集群部署**：多节点通过 gRPC 自动组网，消息跨节点路由转发
- **用户分组**：支持按 `groupId` 对用户分组，可定向推送或全员广播
- **REST API**：提供 HTTP 接口用于发送消息、查询在线用户、集群状态等
- **自动容错**：集群节点间自动握手、节点列表同步、超时用户清理

## 架构

```
┌─────────────────────────────────────────────────────┐
│                    httpPush Node                     │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐ │
│  │ WS /ws   │  │ EX /ex   │  │ REST API          │ │
│  │ WebSocket│  │ HTTP长轮询│  │ /sendmsg /query...│ │
│  └────┬─────┘  └────┬─────┘  └────────┬──────────┘ │
│       └──────────┬──┘                 │             │
│              ┌───▼───┐                │             │
│              │ core  │◄───────────────┘             │
│              │ App   │  用户管理 / 消息路由          │
│              └───┬───┘                              │
│                  │                                   │
│           ┌──────▼──────┐                           │
│           │  cluster    │  gRPC 节点通信             │
│           │  (gRPC)     │  自动组网 / 消息转发       │
│           └─────────────┘                           │
└─────────────────────────────────────────────────────┘
```

### 模块说明

| 模块 | 说明 |
|------|------|
| `main.go` | 入口，加载配置，组装各模块启动 |
| `core` | 全局状态 `App`、消息码头 `MsgDock`、查询参数 |
| `api` | REST API 控制器，消息发送与状态查询 |
| `ws` | WebSocket 连接管理（`/ws`） |
| `ex` | HTTP 长轮询连接管理（`/ex`） |
| `cluster` | 集群 gRPC 通信：握手、节点同步、消息转发、通用查询 |
| `message` | 消息定义（`TextMessage`） |
| `user` | 用户存储、分组管理、历史订单 |
| `util` | 工具集：时间轮、队列、SliceMap、网络参数解析 |

## 快速开始

### 编译

```bash
go build -o httpPush .
```

### 配置

编辑 `config.yml`：

```yaml
web:
  server:
    port: 8084          # HTTP 服务端口
  log:
    level: debug
    file_path: push.log

cluster:
  start: true           # 是否启用集群
  local_port: 8085      # gRPC 端口（默认 HTTP端口+1）
  machine_id: ""        # 节点 ID，留空则自动生成

ex:
  start: true           # 是否启用 HTTP 长轮询
  live_time: 15         # 默认长轮询超时（秒）

ws:
  start: true           # 是否启用 WebSocket
```

### 运行

```bash
./httpPush
```

### 集群部署

在多台机器上分别启动 httpPush，通过 `cluster.remote_host` 配置相邻节点地址即可自动组网：

```yaml
cluster:
  start: true
  local_port: 8085
  remote_host: "192.168.1.2:8085,192.168.1.3:8085"
```

节点间通过 gRPC 自动握手并同步节点列表，消息会跨节点路由到目标用户所在的机器。

## 客户端接入

### WebSocket 连接

```
ws://host:8084/ws?id=用户ID&groupId=分组1,分组2
```

连接建立后，客户端发送 JSON 消息给其他用户：

```json
{"to": "目标用户ID", "msg": "消息内容"}
```

收到推送消息格式：

```json
[{"from": "发送者ID", "body": "消息内容"}]
```

### HTTP 长轮询连接

```
GET /ex?id=用户ID&groupId=分组1,分组2&liveTime=15
```

连接会保持至超时或有消息推送。超时后客户端需重新发起请求。`liveTime` 参数可覆盖默认超时时间。

## REST API

### 发送消息

| 接口 | 参数 | 说明 |
|------|------|------|
| `GET /sendmsg` | `username`/`id`, `msg` | 向指定用户发送文本消息 |
| `GET /sendMessage` | `username`/`id`, `msg` | 发送消息，返回 `{"success": true/false}` |
| `GET /sendGroupMsg` | `groupId`, `msg` | 向分组内所有用户发送消息（`groupId=all` 为全员广播） |

### 查询接口

| 接口 | 参数 | 说明 |
|------|------|------|
| `GET /root_version` | — | 返回版本号和启动时间 |
| `GET /queryUser` | `id`/`username` | 查询用户详细信息（连接地址、时间等） |
| `GET /onlineUser` | — | 返回当前在线用户列表 |
| `GET /info_user` | — | 返回集群各节点信息及总用户数 |
| `GET /queryOrderInfo` | `userId`/`username` | 查询用户在各节点的排序信息 |
| `GET /queryClusterUserNum` | — | 查询集群各节点用户数 |
| `GET /queryGroupInfo` | — | 查询所有分组及成员数 |
| `GET /queryVersion` | — | 查询集群各节点版本和启动时间 |

> 查询接口会自动跨集群节点聚合数据。

## 依赖

- [go-web-frame](https://github.com/chuccp/go-web-frame) — Web 框架
- [gorilla/websocket](https://github.com/gorilla/websocket) — WebSocket 实现
- [google.golang.org/grpc](https://grpc.io/) — gRPC 集群通信
- [go.uber.org/zap](https://github.com/uber-go/zap) — 日志库
