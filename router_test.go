package flyrpc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	uid := int32(123)
	s := Protobuf
	buff, err := s.Marshal(&TestUser{Id: uid, Name: "abc"})
	assert.Nil(t, err)
	r := NewRouter(s)
	// c := make(chan *TestUser, 10)
	var p1, p2, p3, p4 *TestUser
	// inbuff := &bytes.Buffer{}
	// outbuff := &bytes.Buffer{}
	// protocol := newProtocol(inbuff, outbuff, false)
	protocol := NewMockProtocol()
	ctx := NewContext(protocol, r, 0, s)

	r.AddRoute("1", func(u *TestUser) {
		p1 = u
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Cmd:      "1",
		MsgBuff:  buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, uid, p1.Id)

	r.AddRoute("2", func(pkt *Packet, u *TestUser) error {
		assert.Equal(t, "2", pkt.Cmd)
		p2 = u
		return errors.New("e1")
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Cmd:      "2",
		MsgBuff:  buff,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "e1", err.Error())
	assert.Equal(t, uid, p2.Id)

	r.AddRoute("3", func(ctx *Context, u *TestUser) *TestUser {
		p3 = u
		return &TestUser{Id: 567}
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Cmd:      "3",
		MsgBuff:  buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, uid, p3.Id)
	// assert.True(t, len(outbuff.Bytes()) > 0)

	r.AddRoute("4", func(bytes []byte, u *TestUser) (*TestUser, error) {
		u2 := &TestUser{}
		s.Unmarshal(bytes, u2)
		assert.Equal(t, u.Id, u2.Id)
		p4 = u
		return &TestUser{Id: 789}, nil
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Cmd:      "4",
		MsgBuff:  buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, uid, p4.Id)
	// assert.True(t, len(outbuff.Bytes()) > 0)

	r.AddRoute("5", func(ctx *Context, u *TestUser) (*TestUser, error) {
		return nil, NewFlyError(10000)
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Cmd:      "5",
		MsgBuff:  buff,
	})
	assert.Nil(t, err)
	// assert.NotNil(t, err)
	// assert.Equal(t, 10000, err.(*FlyError).Code)

	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Cmd:      "100",
		MsgBuff:  buff,
	})
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.(Error).Code())
	// log.Println(outbuff.Bytes())
}

func TestRouterPanic(t *testing.T) {
	s := Protobuf
	buff, err := s.Marshal(&TestUser{Id: 123, Name: "abc"})
	assert.Nil(t, err)
	r := NewRouter(s)
	protocol := NewMockProtocol()
	ctx := NewContext(protocol, r, 0, s)

	r.AddRoute("1", func(u *TestUser) {
		panic("RouteTest panic")
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Cmd:      "1",
		MsgBuff:  buff,
	})
	assert.Nil(t, err)
}
