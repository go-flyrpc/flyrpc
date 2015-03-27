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
* [OK]自定义
* [OK]json
* [OK]protobuf (proto3)
* [OK]msgpack

## 多路复用
* Gateway Node
* Backend Node

## API

### Server.Listen(addr)

### Server.OnMessage(cmdId, func(*Context, in Message) (out Message, err error))

### Context.SendMessage(cmdId, Message)

### Context.Call(cmdId, Message) (Message, error)

### Client.Connect(addr)

### Client.OnMessage(cmdId, func(*Context, in Message)) (out Message, err error))

### Client.SendMessage(cmdId, Message)

### Client.Call(cmdId, Message) (Message, error)

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

flag|cmdId|msgId|serialized-buff
