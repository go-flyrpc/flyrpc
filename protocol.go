package fly

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
)

const (
	LFLAG_RPC   byte = 0x80
	LFLAG_RESP  byte = 0x40
	LFLAG_ERROR byte = 0x20
	// LFLAG_NOTIFY byte = 0x10
	// LFLAG_LEN_16 byte = 0x08
	// LFLAG_STREAM_MODE byte = 0x04
	// LFLAG_ZIP         byte = 0x02
	// LFLAG_ENCRYPT     byte = 0x01
)

type CmdIdSize uint8
type MsgIdSize uint8
type LengthSize uint16

const MaxLength = ^LengthSize(0)

type Header struct {
	Flag  byte
	CmdId CmdIdSize
	MsgId MsgIdSize
}

type Packet struct {
	Protocol Protocol
	ClientId int
	Length   LengthSize
	Header   *Header
	MsgBuff  []byte
}

func (pkt *Packet) SendPacket(header *Header, bytes []byte) error {
	return pkt.Protocol.SendPacket(&Packet{
		ClientId: pkt.ClientId,
		Header:   header,
		MsgBuff:  bytes,
	})
}

func (pkt *Packet) SendError(err *FlyError) error {
	// TODO
	return nil
}

type PacketHandler func(*Packet)

type Protocol interface {
	OnPacket(PacketHandler)
	SendPacket(*Packet) error
}

type protocol struct {
	Id          int
	IsMultiplex bool
	Conn        net.Conn
	Reader      *bufio.Reader
	Writer      *bufio.Writer
	handlers    []PacketHandler
}

func NewProtocol(conn net.Conn, isMultiplex bool) Protocol {
	return newProtocol(conn, conn, isMultiplex)
}

func newProtocol(reader io.Reader, writer io.Writer, isMultiplex bool) Protocol {
	p := &protocol{
		IsMultiplex: isMultiplex,
		Reader:      bufio.NewReader(reader),
		Writer:      bufio.NewWriter(writer),
		handlers:    make([]PacketHandler, 0),
	}
	go p.handleStream()
	return p
}

func (p *protocol) Close() {
	if p.Conn != nil {
		p.Conn.Close()
	}
}

func (p *protocol) OnPacket(handler PacketHandler) {
	p.handlers = append(p.handlers, handler)
}

func (p *protocol) SendPacket(pk *Packet) error {
	if p.Writer == nil {
		p.Close()
		return NewFlyError(ERR_WRITER_CLOSED, nil)
	}
	if p.Writer.Available() == 0 {
		return NewFlyError(ERR_WRITER_CLOSED, nil)
	}
	if p.IsMultiplex {
		binary.Write(p.Writer, binary.BigEndian, pk.ClientId)
	}
	// if support zip {
	// DO zip
	// buff = zip(buff)
	// }
	pk.Length = LengthSize(len(pk.MsgBuff))
	if pk.Length > MaxLength {
		return NewFlyError(ERR_BUFF_TO_LONG, nil)
	}
	binary.Write(p.Writer, binary.BigEndian, pk.Header)

	binary.Write(p.Writer, binary.BigEndian, pk.Length)
	p.Writer.Write(pk.MsgBuff)
	p.Writer.Flush()
	return nil
}

func (p *protocol) handleStream() {
	// 协议处理函数
	for {
		log.Println("read packet")
		msg, err := p.ReadPacket()
		log.Println("packet", msg)
		if err != nil {
			if err != io.EOF {
				log.Println("error", err)
			}
			p.Close()
			break
		}
		for _, handler := range p.handlers {
			log.Println("handle msg")
			go handler(msg)
		}
	}
}

func (p *protocol) ReadPacket() (*Packet, error) {
	log.Println("reading packet")
	var clientId = p.Id
	if p.IsMultiplex {
		log.Println("reading packet 1")
		err := binary.Read(p.Reader, binary.BigEndian, &clientId)
		if err != nil {
			return nil, err
		}
	}
	log.Println("reading header")
	// read header
	header := &Header{}
	err := binary.Read(p.Reader, binary.BigEndian, header)
	if err != nil {
		return nil, err
	}
	log.Println("reading Length ")
	// read length
	var length LengthSize
	err = binary.Read(p.Reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	log.Println("length is", length)
	log.Println("reading buff")
	buf := make([]byte, length)
	_, err = io.ReadFull(p.Reader, buf)
	if err != nil {
		return nil, err
	}
	log.Println("return packet", buf)
	packet := &Packet{
		ClientId: clientId,
		Length:   length,
		Header:   header,
		MsgBuff:  buf,
	}
	return packet, nil
}
