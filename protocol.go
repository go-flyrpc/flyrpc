package fly

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
)

// TypeBits - bits of sub protocol
// TypeRPC  - type of RPC. Main feature
// TypePing - type of Ping. Keepalive
// TypeHello - type of Hello. Tell the client information related with protocol, like version, zip, supported encoding
const (
	FlagBitsType byte = 0xb0 // 11000000
	TypeRPC      byte = 0xb0 // 11000000
	TypePing     byte = 0x80 // 10000000
	TypeHello    byte = 0x40 // 01000000
	TypeMessage  byte = 0x00 // 00000000
)

const (
	RPCFlagResp   byte = 0x01
	RPCFlagError  byte = 0x02
	RPCFlagBuffer byte = 0x04
	PingFlagPing  byte = 0x01
	PingFlagPong  byte = 0x02
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
	SubType  byte
	Length   TLength
	Header   *Header
	MsgBuff  []byte
}

type Protocol interface {
	ReadPacket() (*Packet, error)
	SendPacket(*Packet) error
	Close() error
}

type protocol struct {
	IsMultiplex bool
	Conn        net.Conn
	Reader      *bufio.Reader
	Writer      *bufio.Writer
}

func NewProtocol(conn net.Conn, isMultiplex bool) Protocol {
	protocol := newProtocol(conn, conn, isMultiplex)
	protocol.Conn = conn
	return protocol
}

func newProtocol(reader io.Reader, writer io.Writer, isMultiplex bool) *protocol {
	p := &protocol{
		IsMultiplex: isMultiplex,
		Reader:      bufio.NewReader(reader),
		Writer:      bufio.NewWriter(writer),
	}
	return p
}

func (p *protocol) Close() error {
	if p.Conn != nil {
		return p.Conn.Close()
	}
	return nil
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
	subType := header.Flag & FlagBitsType
	if subType == TypePing {
		return p.readPingPacket(header, clientId)
	}
	if subType == TypeHello {
		return p.readHelloPacket(header, clientId)
	}
	// TypeRPC | TypeMessage
	return p.readRPCPacket(header, clientId)
}

func (p *protocol) readRPCPacket(header *Header, clientId int) (*Packet, error) {
	// read length
	var length TLength
	err := binary.Read(p.Reader, binary.BigEndian, &length)
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
		SubType:  TypeRPC,
		Length:   length,
		Header:   header,
		MsgBuff:  buf,
	}
	return packet, nil
}

func (p *protocol) readPingPacket(header *Header, clientId int) (*Packet, error) {
	// TODO support sized ping packet
	return &Packet{
		ClientId: clientId,
		SubType:  TypePing,
		Length:   0,
		Header:   header,
		MsgBuff:  nil,
	}, nil
}

func (p *protocol) readHelloPacket(header *Header, clientId int) (*Packet, error) {
	// TODO Well design hello protocol
	return &Packet{
		ClientId: clientId,
		SubType:  TypeHello,
		Length:   0,
		Header:   header,
		MsgBuff:  nil,
	}, nil
}
