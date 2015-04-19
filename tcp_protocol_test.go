package fly

import (
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocolReal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:17777")
	assert.Nil(t, err)

	conn1, err1 := net.Dial("tcp", "127.0.0.1:17777")
	assert.Nil(t, err1)
	p1 := NewTcpProtocol(conn1, false)

	conn2, err2 := listener.Accept()
	assert.Nil(t, err2)
	p2 := NewTcpProtocol(conn2, false)

	flag := TypeRPC | RPCFlagResp | RPCFlagError

	err = p1.SendPacket(&Packet{
		Header: &Header{
			Flag: flag,
			Cmd:  222,
			Seq:  123,
		},
		MsgBuff: []byte{1, 2, 3, 4, 5, 6},
	})
	assert.Nil(t, err)

	pkt, err := p2.ReadPacket()

	log.Println("return packet", pkt.MsgBuff)
	assert.Equal(t, TCmd(flag), pkt.Header.Flag)
	assert.Equal(t, TCmd(222), pkt.Header.Cmd)
	assert.Equal(t, TCmd(123), pkt.Header.Seq)
	assert.Equal(t, 6, len(pkt.MsgBuff))
	assert.Equal(t, byte(6), pkt.MsgBuff[5])
	err = conn1.Close()
	assert.Nil(t, err)
}
