package fly

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextSendMessage(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	protocol := new(MockProtocol)
	serializer := Protobuf
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)
	protocol.OnPacket(context.emitPacket)
	c := make(chan *TestUser, 1)
	router.AddRoute(1, func(ctx *Packet, in *TestUser) {
		c <- in
	})
	err := context.SendMessage(1, &TestUser{Id: 123})
	assert.Nil(t, err)
	u := <-c
	assert.Equal(t, 123, u.Id)
}

func TestContextCall(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	protocol := new(MockProtocol)
	serializer := Protobuf
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)
	protocol.OnPacket(context.emitPacket)
	router.AddRoute(1, func(ctx *Packet, in *TestUser) *TestUser {
		return &TestUser{Id: in.Id + 1}
	})
	var reply = new(TestUser)
	err := context.Call(1, reply, &TestUser{Id: 123})
	assert.Nil(t, err)
	assert.Equal(t, 124, reply.Id)
}
