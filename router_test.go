package fly

import (
	"bytes"
	"errors"
	"log"
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
	r.AddRoute(1, func(p *Packet, u *TestUser) {
		p1 = u
	})
	r.AddRoute(2, func(p *Packet, u *TestUser) error {
		p2 = u
		return errors.New("e1")
	})
	r.AddRoute(3, func(p *Packet, u *TestUser) *TestUser {
		p3 = u
		return &TestUser{Id: 567}
	})
	r.AddRoute(4, func(p *Packet, u *TestUser) (*TestUser, error) {
		p4 = u
		return &TestUser{Id: 789}, nil
	})
	r.AddRoute(5, func(p *Packet, u *TestUser) (*TestUser, error) {
		return nil, NewFlyError(10000)
	})
	inbuff := &bytes.Buffer{}
	outbuff := &bytes.Buffer{}
	protocol := newProtocol(inbuff, outbuff, false)

	err = r.onPacket(nil, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 1,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, 123, p1.Id)

	err = r.onPacket(nil, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 2,
		},
		MsgBuff: buff,
	})
	assert.NotNil(t, err)
	assert.Equal(t, "e1", err.Error())
	assert.Equal(t, 123, p2.Id)

	err = r.onPacket(nil, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 3,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, 123, p3.Id)
	assert.True(t, len(outbuff.Bytes()) > 0)

	err = r.onPacket(nil, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 4,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	assert.Equal(t, 123, p4.Id)
	assert.True(t, len(outbuff.Bytes()) > 0)

	err = r.onPacket(nil, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 5,
		},
		MsgBuff: buff,
	})
	assert.Nil(t, err)
	// assert.NotNil(t, err)
	// assert.Equal(t, 10000, err.(*FlyError).Code)

	err = r.onPacket(nil, &Packet{
		Protocol: protocol,
		Header: &Header{
			CmdId: 100,
		},
		MsgBuff: buff,
	})
	assert.NotNil(t, err)
	assert.Equal(t, ERR_NOT_FOUND, err.(*FlyError).Code)
	log.Println(outbuff.Bytes())
}
