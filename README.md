# 草案
模式
* [OK]Send/Recv
* [OK]Req/Res
* Pub/Sub

网络协议 
* [OK]TCP
* UDP
* Websocket
* P2P

序列化接口 
* [OK]自定义
* [OK]json
* [OK]protobuf (proto3)
* [OK]msgpack

多路复用
* Gateway Node
* Backend Node

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

## TODO 多路复用协议
length|buff
clientId | length | buff
worker
线程安全，防止两个客户端连接写同步

## 消息协议
flag|cmdId|msgId|buff

##
SendMessage / OnMessage

## high level base on SendMessage and OnMessage
SendCall  / OnCall
Publish / Subscribe / OnChannelMessage
