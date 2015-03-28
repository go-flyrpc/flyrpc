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
	p1 := NewProtocol(conn1, false)

	conn2, err2 := listener.Accept()
	assert.Nil(t, err2)
	p2 := NewProtocol(conn2, false)

	cpkt := make(chan *Packet, 1)
	p2.OnPacket(func(pkt *Packet) {
		cpkt <- pkt
	})

	p1.SendPacket(&Packet{
		Header: &Header{
			Flag:  111,
			CmdId: 222,
			MsgId: 123,
		},
		MsgBuff: []byte{1, 2, 3, 4, 5, 6},
	})

	pkt := <-cpkt
	log.Println("return packet", pkt.MsgBuff)
	assert.Equal(t, CmdIdSize(111), pkt.Header.Flag)
	assert.Equal(t, CmdIdSize(222), pkt.Header.CmdId)
	assert.Equal(t, CmdIdSize(123), pkt.Header.MsgId)
	assert.Equal(t, 6, len(pkt.MsgBuff))
	assert.Equal(t, byte(6), pkt.MsgBuff[5])
}

/*
func TestProtocol(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	buff := &bytes.Buffer{}
	p := newProtocol(buff, buff, false)
	c := make(chan *Packet)
	p.OnPacket(func(packet *Packet) {
		c <- packet
	})
	p.SendPacket(&Packet{
		Header: &Header{
			Flag:  0x00,
			CmdId: 200,
			MsgId: 11,
		},
		MsgBuff: []byte{1, 2, 3, 4, 5, 6},
	})
	buff.Write([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
	// go p.handleStream()
	packet := <-c
	log.Println("test packet:", packet.Header, packet.MsgBuff, packet.Length)
}
*/

type MockProtocol struct {
	handlers []PacketHandler
}

func (mp *MockProtocol) SendPacket(pkt *Packet) error {
	for _, handler := range mp.handlers {
		go handler(pkt)
	}
	return nil
}

func (mp *MockProtocol) OnPacket(handler PacketHandler) {
	mp.handlers = append(mp.handlers, handler)
}
