package fly

import (
	"bytes"
	"log"
	"testing"
)

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
	packet := <-c
	t.Log(packet.Header, packet.MsgBuff, packet.Length)
	t.Log("Hello")
}
