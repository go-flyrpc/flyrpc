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
MessageHandler 可以有以下几种形式

```go
// 有返回值，用于处理Call
func(*Context, in Message) (out Message, err error)
func(*Context, in Message) out Message

// 无返回值，用于处理SendMessage
func(*Context, in Message) err error
func(*Context, in Message)
```

#### Server.Listen(addr)

#### Server.OnMessage(cmd, MessageHandler)

#### Context.SendMessage(cmd, Message)

#### Context.Call(cmd, Message) (Message, error)

#### Client.Connect(addr)

#### Client.OnMessage(cmd, MessageHandler)

#### Client.SendMessage(cmd, Message)

#### Client.Call(cmd, Message) (Message, error)

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

# 协议
## 分包协议
length|buff

## TODO 多路复用协议
clientIds | length | buff

NOTICE:
线程安全，防止两个客户端连接写同步

## 消息协议

flag|cmd|seq|serialized-buff
