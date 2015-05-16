
[![Build Status](https://travis-ci.org/flyrpc/flyrpc.svg?branch=master)](https://travis-ci.org/flyrpc/flyrpc)
[![Coverage Status](https://coveralls.io/repos/flyrpc/flyrpc/badge.svg?branch=master)](https://coveralls.io/r/flyrpc/flyrpc?branch=master)


FlyRPC是适合高频率通信的网络框架。

# 协议

## 消息协议

| Flag   | Command   | Sequence  | Buffer Length | Buffer |
| ------ |:---------:| ---------:|:-------------:| ------ |
| 1 byte | 2 byte    | 1 byte    | 2 byte        | n byte |
| 标志位 | 命令ID    | 序列ID    | 消息长度      | 消息体 |

### Flag说明

| 子协议 | 控制位 |
| -----: | ------ |
| 2 bit  | 6 bit  |

| 子协议 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
| ------ |---|---|---|---|---|---|---|---|
| RPC    | 1 | 1 | ? | ? | ? | Buffer | Error | Resp |
| Ping   | 1 | 0 | ? | ? | ? | ? | Pong | Ping |
| Helo   | 0 | 1 | ? | ? | ? | ? | ? | ? |
| MQ     | 0 | 0 | ? | ? | ? | ? | ? | ? |

## 服务器内部多路复用协议

| Client count  | Client Id 1   | ...  | Client Id n | Buffer Length | Buffer |
| ------------- |:-------------:| ----:|:-----------:| ------------- | ------ |
| 1 byte        | 2 byte        | ...  | 2 byte      | 2byte         | n byte |

# 草案
## 模式
* [OK]Send/Recv
* [OK]Req/Res
* Pub/Sub

## 网络协议 
* [OK]TCP
* UDP
* Websocket
* P2P

## 序列化接口 
* 数据压缩
* [OK]自定义
* [OK]json
* [OK]protobuf (proto3)
* [OK]msgpack

## 多路复用
* Gateway Node
* Backend Node

## API

#### type MessageHandler
MessageHandler 可以有以下几种形式的参数的组合或无参数
* \*Context
* \*Packet 
* \[]byte
* \*UserCustomMessage

返回可以是：
* UserCustomMessage
* UserCustomMessage, error
* non return
* error


```go
// 有返回值，用于处理Call
func(*Context, in MyMessage) (out Message, err error)
func(*Packet, in MyMessage) out Message

// 无返回值，用于处理SendMessage
func(bytes []byte) err error
func()
```

#### Server.Listen(addr)

#### Server.OnMessage(cmd, MessageHandler)

#### Context.SendMessage(cmd, Message)

#### Context.Call(cmd, Message) (Message, error)

#### Context.Ping(length, timeout) error

#### Client.Connect(addr)

#### Client.OnMessage(cmd, MessageHandler)

#### Client.SendMessage(cmd, Message)

#### Client.Call(cmd, Message) (Message, error)

#### Client.Ping(length, timeout) error

## 待定
```
type ClientStub struct {
    foo func(a) b `flyid:1`
}

rpc := &ClientStub{}
client.InjectService(rpc)
b := rpc.foo(a)
```

# 类关联结构
```
TCP/UDP/WS        Packet    json/protobuf/msgpack
 |                + + +                   |
 |                | | |                   |
 -->Protocol ------ | |      Serializer <--
      +       ------- |          +
      |       | ----Route --------
      |       | |     *+
      |       | |     |
      |       | |   Router
      |       | |    + +
      ------- | | ---- |
          1*| | + |    |
       --->Context     |
       |transport+*    |
       |        ------ |
       |             | |
 -->Client          Server<--   MultiplexedServer<--
 |                          |                      |
 |                          |                      |
TCP/UDP/WS            TCP/UDP/WS               TCP/UDP/WS
```
* _\*_ 多实例
* _-->_ 继承或实现
* _\+_  被引用
