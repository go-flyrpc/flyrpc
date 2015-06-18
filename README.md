
[![Build Status](https://travis-ci.org/flyrpc/flyrpc.svg?branch=master)](https://travis-ci.org/flyrpc/flyrpc)
[![Coverage Status](https://coveralls.io/repos/flyrpc/flyrpc/badge.svg?branch=master)](https://coveralls.io/r/flyrpc/flyrpc?branch=master)

FlyRPC is a PROTOCOL implements maximum features with minimal packet size.

```
go get gopkg.in/flyrpc.v1
```

# Protocol

## Message Protocol

| Packet Length | Flag   | Transfer Flag | Sequence  | Command   | Buffer  | CRC16   |
|:-------------:| ------ | ------------- | ---------:|:---------:| ------- | ------- |
| 2 bytes       | 1 byte | 1 byte        | 2 bytes   | string\n  | n bytes | 2 bytes |

### Flag Spec

| SubType | Options |
| ------: | ------- |
| 2 bit   | 6 bit   |

| SubType | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
| ------- |---|---|---|---|---|---|---|---|
| RPC     | 1 | 1 | ? | CRC16 | Buffer | Error | Resp | Req |
| Ping    | 1 | 0 | ? | CRC16 | ? | ? | Pong | Ping |
| Helo    | 0 | 1 | ? | CRC16 | ? | ? | ? | ? |
| MQ      | 0 | 0 | ? | CRC16 | ? | ? | ? | ? |

CRC16 allow to add salt.

_NOTE_ CRC16 not support yet, keep it 0.

### Transfer Flag

Transfer Flag 0x00 should be `plain` `json`.

_NOTE_ Not complete.

Design A

| Accept Encoding | Encoding | Accept Serializer | Serializer |
|:--------------- | -------- | ----------------- | ---------- |
| 2 bits          | 2 bits   | 2 bits            | 2 bits     |

Up to 4 encoding and 4 serializer.

Design B

| Encoding | Serializer |
| -------- | ---------- |
| 2 bits   | 2 bits     |

If HELO, it means Accept-Encoding and Accept-Serializer

Up to 16 encoding and 16 serializer. (Who need 16 encoding and 16 serializer)

### Buffer

If `Error` flag is `1`, buffer is `string`, it is the error code.

If `Error` flag is `0`, buffer is the serialized bytes.

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
