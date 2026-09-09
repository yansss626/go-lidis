# go-lidis

LIDIS KV 存储系统的 Go 客户端 SDK。

`go-lidis` 用于在 Go 应用程序中连接并操作 LIDIS KV 存储服务，
提供连接池管理、自定义协议封装以及响应解析功能。

LIDIS 服务端采用 C 语言实现，本项目提供对应的 Go 客户端访问接口。

## 特性

- 基于 TCP 协议通信
- 支持连接池管理
- 支持自定义 LIDIS 通信协议
- 支持命令封装与响应解析
- 提供简单易用的 Go API
- 轻量级，无第三方服务依赖

---

## 环境要求

- Go 1.20+

- 已启动 LIDIS KV 存储服务

LIDIS 服务端：

https://github.com/yansss626/lidis

---

## 安装

使用 `go get` 安装：

```bash
go get github.com/yansss626/go-lidis
```

导入：

```go
import lidis "github.com/yansss626/go-lidis"
```

---

# 快速开始

## 创建客户端

```go
package main

import (
    "fmt"

    lidis "github.com/yansss626/go-lidis"
)

func main() {

    config := lidis.Config{
        Host:        "127.0.0.1",
        Port:        8080,
        InitialCap:  5,
        MaxIdle:     10,
        MaxCap:      20,
        IdleTimeout: 60,
    }

    pool, err := lidis.NewConnectionPool(config)
    if err != nil {
        panic(err)
    }

    client := lidis.NewClient(pool)

    fmt.Println("LIDIS client created")

    _ = client
}
```

# 工作流程

```
        Go 应用程序
              |
              |
        go-lidis Client
              |
              |
       TCP Connection Pool
              |
              |
        LIDIS Server
        (C语言实现)
```

每个连接包含：

```
Connection

 ├── net.Conn
 └── bufio.Reader
```

客户端通过连接池管理 TCP 连接，并完成：

- 请求发送
- 响应接收
- 协议解析

---

# 项目结构

```
go-lidis
|
├── client.go       # 客户端与连接池管理
├── command.go      # LIDIS协议命令封装
├── reply.go        # 服务端响应解析
├── go.mod
└── README.md
```

---

# 相关项目

LIDIS KV 存储服务端：

https://github.com/yansss626/lidis

---

# License

MIT License
