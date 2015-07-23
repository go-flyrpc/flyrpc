
[![Build Status](https://travis-ci.org/flyrpc/flyrpc.svg?branch=master)](https://travis-ci.org/flyrpc/flyrpc)
[![Coverage Status](https://coveralls.io/repos/flyrpc/flyrpc/badge.svg?branch=master)](https://coveralls.io/r/flyrpc/flyrpc?branch=master)

FlyRPC is a PROTOCOL implements maximum features with minimal packet size.

```
go get gopkg.in/flyrpc.v1
```

# Protocol

## Message Protocol

|Name | Flag   | Length | Sequence | CMD(only Req) | Buffer  |
|-----|:------:|:------:|:--------:|:------:|:-------:|
|Bytes| 1      | 1-4    | 2        |string\0| *       |

### Flag Spec

| 1     | 2       | 3 | 4     | 5 | 6 | 7 - 8        |
|-------|---------|---|-------|---|---|--------------|
|Req 0  |Must ACK |Zip|Partial|   |   | length bytes |
|Res 1  |Error    |   |       |   |   |              |

### Buffer

If `Error` flag is `1`, buffer is `string`, it is the error code.

If `Error` flag is `0`, buffer is the serialized bytes wanted type.

Support `[]byte` `string` `struct` `map`

Not support yet `int` `float`

## Internal Multiplexed Protocol

| Client count  | Client Id 1   | ...  | Client Id n | Buffer Length | Buffer |
| ------------- |:-------------:| ----:|:-----------:| ------------- | ------ |
| 1 byte        | 2 byte        | ...  | 2 byte      | 2byte         | n byte |

# Draft
## Patterns
* [OK]Send/Recv
* [OK]Req/Res
* Pub/Sub

## Network
* [OK]TCP
* UDP
* Websocket
* P2P

## Serializer
* Compress
* [OK]json
* [OK]protobuf (proto3)
* [OK]msgpack

## Multiplexing
* Gateway Node
* Backend Node

## API

#### type MessageHandler
MessageHandler could take below params
* \*Context
* \*Packet 
* \[]byte
* \*UserCustomMessage

MessageHandler could return below results
* UserCustomMessage, error
* UserCustomMessage
* error
* no return

#### NewServer(*ServerOpts) *Server

#### Server.Listen(addr)

#### Server.OnMessage(cmd, MessageHandler)

#### Context.SendMessage(cmd, Message)

#### Context.Call(cmd, Message) (Message, error)

#### Context.Ping(length, timeout) error

#### NewClient(addr) *Client

#### Client.Connect(addr)

#### Client.OnMessage(cmd, MessageHandler)

#### Client.SendMessage(cmd, Message)

#### Client.Call(cmd, Message) (Message, error)

#### Client.Ping(length, timeout) error

# Class Digrame
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
* _\*_ Multiple instance
* _-->_ Extends
* _\+_  Reference
