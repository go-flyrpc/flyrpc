package fly

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	s := Protobuf
	buff, err := s.Marshal(&TestUser{Id: 123, Name: "abc"})
	assert.Nil(t, err)
	r := NewRouter(s)
	// c := make(chan *TestUser, 10)
	var p1, p2, p3, p4 *TestUser
	// inbuff := &bytes.Buffer{}
	// outbuff := &bytes.Buffer{}
	// protocol := newProtocol(inbuff, outbuff, false)
	protocol := new(MockProtocol)
	ctx := NewContext(protocol, r, 0, s)

	r.AddRoute(1, func(ctx *Context, u *TestUser) {
		p1 = u
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 1,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, 123, p1.Id)

	r.AddRoute(2, func(ctx *Context, u *TestUser) error {
		p2 = u
		return errors.New("e1")
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 2,
		},
		MsgBuff: buff,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "e1", err.Error())
	assert.Equal(t, 123, p2.Id)

	r.AddRoute(3, func(ctx *Context, u *TestUser) *TestUser {
		p3 = u
		return &TestUser{Id: 567}
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 3,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, 123, p3.Id)
	// assert.True(t, len(outbuff.Bytes()) > 0)

	r.AddRoute(4, func(ctx *Context, u *TestUser) (*TestUser, error) {
		p4 = u
		return &TestUser{Id: 789}, nil
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 4,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, 123, p4.Id)
	// assert.True(t, len(outbuff.Bytes()) > 0)

	r.AddRoute(5, func(ctx *Context, u *TestUser) (*TestUser, error) {
		return nil, NewFlyError(10000)
	})
	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 5,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	// assert.NotNil(t, err)
	// assert.Equal(t, 10000, err.(*FlyError).Code)

	err = r.emitPacket(ctx, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 100,
		},
		MsgBuff: buff,
	})
	assert.NotNil(t, err)
	assert.Equal(t, ERR_NOT_FOUND, err.(*FlyError).Code)
	// log.Println(outbuff.Bytes())
}
