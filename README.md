
[![Build Status](https://travis-ci.org/flyrpc/flyrpc.svg?branch=master)](https://travis-ci.org/flyrpc/flyrpc)
[![Coverage Status](https://coveralls.io/repos/flyrpc/flyrpc/badge.svg?branch=master)](https://coveralls.io/r/flyrpc/flyrpc?branch=master)


FlyRPC high speed flexible network framework.

# Protocol

## Message Protocol

| Flag   | Command   | Sequence  | Buffer Length | Buffer |
| ------ |:---------:| ---------:|:-------------:| ------ |
| 1 byte | 2 byte    | 1 byte    | 2 byte        | n byte |

### Flag Spec

| SubType | Options |
| ------: | ------- |
| 2 bit   | 6 bit   |

| SubType | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
| ------- |---|---|---|---|---|---|---|---|
| RPC     | 1 | 1 | ? | ? | ? | Buffer | Error | Resp |
| Ping    | 1 | 0 | ? | ? | ? | ? | Pong | Ping |
| Helo    | 0 | 1 | ? | ? | ? | ? | ? | ? |
| MQ      | 0 | 0 | ? | ? | ? | ? | ? | ? |

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
