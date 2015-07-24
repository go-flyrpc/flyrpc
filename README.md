
[![Build Status](https://travis-ci.org/flyrpc/flyrpc.svg?branch=master)](https://travis-ci.org/flyrpc/flyrpc)
[![Coverage Status](https://coveralls.io/repos/flyrpc/flyrpc/badge.svg?branch=master)](https://coveralls.io/r/flyrpc/flyrpc?branch=master)

FlyRPC is a PROTOCOL implements maximum features with minimal packet size.

```
go get gopkg.in/flyrpc.v1
```

# What the protocol defined.

* Asynchronous Request/Response
* Request with unlimited string code and unlimited binary payload.
* Response with unlimited string code and unlimited binary payload.
* Compress code and payload
* Request can have an implicit `ack response` or `no response`.

# Protocol

## Packet Spec

|Name   | Flag   | Sequence |Code    | Length | Payload |
|-------|:------:|:--------:|:------:|:------:|:-------:|
|Bytes  | 1      | 2        |string\0| 1,2,4,8| *       |

### Flag Spec

| 1      | 2           | 3 | 4 | 5      | 6         | 7 - 8        |
|--------|-------------|---|---|--------|-----------|--------------|
|Response|Wait Response|   |   |Zip Code|Zip Payload| length bytes |

# API

```js
// level 1
conn.onRawPacket(flag, seq, rawCode, rawPayload)
conn.sendRawPacket(flag, seq, rawCode, rawPayload)

// level 2
conn.onPacket(flag, seq, code, payload)
conn.sendPacket(flag, seq, code, payload)

// level 3
conn.sendRequest(code, payload)
conn.sendResponse(seq, code, payload)
conn.onRequest(function handler(seq, code, payload))
conn.onResponse(function handler(seq, code, payload))

// level 4
conn.send(code, payload)
conn.request(code, payload, function callback(err, code, payload))
conn.handle(code, function handler(payload, function reply(code, payload)))
```

### conn.OnMessage(func(code, payload) (code, payload))
### conn.Request(code, payload) (code, payload)
### conn.Send(code, payload)

# Draft

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

#### Server.OnMessage(path, MessageHandler)

#### Context.SendMessage(path, Message)

#### Context.Call(path, Message) (Message, error)

#### Context.Ping(length, timeout) error

#### NewClient(addr) *Client

#### Client.Connect(addr)

#### Client.OnMessage(path, MessageHandler)

#### Client.SendMessage(path, Message)

#### Client.Call(path, Message) (Message, error)

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
