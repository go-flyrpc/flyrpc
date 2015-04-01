package fly

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
)

const (
	LFlagRPC    byte = 0x80
	LFlagResp   byte = 0x40
	LFlagError  byte = 0x20
	LFlagBuffer byte = 0x10
	// LFLAG_NOTIFY byte = 0x10
	// LFLAG_LEN_16 byte = 0x08
	// LFLAG_ZIP         byte = 0x02
	// LFLAG_ENCRYPT     byte = 0x01
)

type TCmd uint16
type TSeq uint8
type TLength uint16

const MaxLength = ^TLength(0)

type Header struct {
	Flag byte
	Cmd  TCmd
	Seq  TSeq
}

type Packet struct {
	// TODO remove this from packet
	Protocol Protocol
	ClientId int
	Length   TLength
	Header   *Header
	MsgBuff  []byte
}

type PacketHandler func(*Packet)

type Protocol interface {
	OnPacket(PacketHandler)
	SendPacket(*Packet) error
}

type protocol struct {
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

func (p *protocol) Close() error {
	if p.Conn != nil {
		return p.Conn.Close()
	}
	return nil
}

func (p *protocol) OnPacket(handler PacketHandler) {
	p.handlers = append(p.handlers, handler)
}

func (p *protocol) SendPacket(pk *Packet) error {
	// FIXME length, header position wrong.
	log.Println("Sending:", pk.ClientId, pk.Header, pk.MsgBuff)
	if p.Writer == nil {
		err := p.Close()
		return NewFlyError(ErrWriterClosed, err)
	}
	if p.Writer.Available() == 0 {
		return NewFlyError(ErrWriterClosed, nil)
	}
	if p.IsMultiplex {
		if err := binary.Write(p.Writer, binary.BigEndian, pk.ClientId); err != nil {
			return err
		}
	}
	// if support zip {
	// DO zip
	// buff = zip(buff)
	// }
	pk.Length = TLength(len(pk.MsgBuff))
	if pk.Length > MaxLength {
		return NewFlyError(ErrBuffTooLong, nil)
	}
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Header); err != nil {
		return err
	}
	// log.Println("Length:", pk.Length)

	if err := binary.Write(p.Writer, binary.BigEndian, pk.Length); err != nil {
		return err
	}
	log.Println("Buff:", pk.MsgBuff)
	if _, err := p.Writer.Write(pk.MsgBuff); err != nil {
		return err
	}
	return p.Writer.Flush()
}

func (p *protocol) handleStream() {
	// 协议处理函数
	for {
		msg, err := p.ReadPacket()
		log.Println("readPacket:", msg, err)
		if err != nil {
			if err != io.EOF {
				log.Println("error", err)
			}
			_ = p.Close()
			break
		}
		// emitPacket
		for _, handler := range p.handlers {
			go handler(msg)
		}
	}
}

func (p *protocol) ReadPacket() (*Packet, error) {
	// FIXME length, header position wrong.
	var clientId = 0
	// only for server
	if p.IsMultiplex {
		err := binary.Read(p.Reader, binary.BigEndian, &clientId)
		if err != nil {
			return nil, err
		}
	}
	// read header
	header := &Header{}
	err := binary.Read(p.Reader, binary.BigEndian, header)
	if err != nil {
		return nil, err
	}
	// read length
	var length TLength
	err = binary.Read(p.Reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(p.Reader, buf)
	if err != nil {
		return nil, err
	}
	// log.Println("return packet", buf)
	packet := &Packet{
		ClientId: clientId,
		Length:   length,
		Header:   header,
		MsgBuff:  buf,
	}
	return packet, nil
}
