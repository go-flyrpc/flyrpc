package flyrpc

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextSendMessage(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	protocol := NewMockProtocol()
	serializer := Protobuf
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)
	uid := int32(123)

	c := make(chan *TestUser, 2)
	router.AddRoute(1, func(ctx *Context, in *TestUser) {
		c <- in
	})
	go func() {
		err := context.SendMessage(1, &TestUser{Id: uid})
		assert.Nil(t, err)
	}()
	go func() {
		err := context.SendMessage(1, &TestUser{Id: uid})
		assert.Nil(t, err)
	}()
	pkt, err := protocol.ReadPacket()
	assert.Nil(t, err)
	context.emitPacket(pkt)
	pkt, err = protocol.ReadPacket()
	assert.Nil(t, err)
	context.emitPacket(pkt)
	u := <-c
	assert.Equal(t, uid, u.Id)
	u = <-c
	assert.Equal(t, uid, u.Id)
}

func TestContextCall(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	protocol := NewMockDelayProtocol(50 * time.Millisecond)
	serializer := Protobuf
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)
	router.AddRoute(1, func(ctx *Context, in *TestUser) *TestUser {
		return &TestUser{Id: in.Id + 1}
	})
	var reply = new(TestUser)
	go func() {
		for {
			pkt, err := protocol.ReadPacket()
			context.emitPacket(pkt)
			if err != nil {
				break
			}
		}
	}()
	err := context.Call(1, reply, &TestUser{Id: 123})
	assert.Nil(t, err)
	assert.Equal(t, int32(124), reply.Id)

	context.timeout = time.Millisecond
	err = context.Call(1, reply, &TestUser{Id: 123})
	assert.Error(t, err)
	assert.Equal(t, ErrTimeOut, err.(Error).Code())
}

func TestCallAck(t *testing.T) {
	protocol := NewMockDelayProtocol(time.Millisecond)
	serializer := Protobuf
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)
	router.AddRoute(1, func(ctx *Context, in *TestUser) {
	})
	go func() {
		for {
			pkt, err := protocol.ReadPacket()
			context.emitPacket(pkt)
			if err != nil {
				break
			}
		}
	}()
	context.timeout = 200 * time.Millisecond
	err := context.Call(1, nil, &TestUser{Id: 123})
	assert.NoError(t, err)
}

func TestCallTimeout(t *testing.T) {
	protocol := NewMockDelayProtocol(time.Second)
	serializer := Protobuf
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)
	router.AddRoute(1, func(ctx *Context, in *TestUser) *TestUser {
		return &TestUser{Id: in.Id + 1}
	})
	var reply = new(TestUser)
	go func() {
		for {
			pkt, err := protocol.ReadPacket()
			context.emitPacket(pkt)
			if err != nil {
				break
			}
		}
	}()
	context.timeout = 10 * time.Millisecond
	err := context.Call(1, reply, &TestUser{Id: 123})
	assert.Error(t, err)
	assert.Equal(t, ErrTimeOut, err.(Error).Code())
}

func TestPing(t *testing.T) {
	protocol := NewMockDelayProtocol(50 * time.Millisecond)
	serializer := Protobuf
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)

	go func() {
		for {
			pkt, err := protocol.ReadPacket()
			context.emitPacket(pkt)
			if err != nil {
				break
			}
		}
	}()

	err := context.Ping(0, 200*time.Millisecond)
	assert.NoError(t, err)

	/*
		err = context.Ping(0, 10*time.Millisecond)
		assert.Error(t, err)
		assert.Equal(t, ErrTimeOut, err.(*flyError).Code)
	*/
}
