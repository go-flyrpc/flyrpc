package flyrpc

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"reflect"
)

type TcpProtocol struct {
	IsMultiplex bool
	Conn        net.Conn
	Reader      *bufio.Reader
	Writer      *bufio.Writer
}

func NewTcpProtocol(conn net.Conn, isMultiplex bool) *TcpProtocol {
	if conn == nil || reflect.ValueOf(conn).IsNil() {
		panic("conn should not be nil")
	}
	protocol := newTcpProtocol(conn, conn, isMultiplex)
	protocol.Conn = conn
	return protocol
}

func newTcpProtocol(reader io.Reader, writer io.Writer, isMultiplex bool) *TcpProtocol {
	p := &TcpProtocol{
		IsMultiplex: isMultiplex,
		Reader:      bufio.NewReader(reader),
		Writer:      bufio.NewWriter(writer),
	}
	return p
}

func (p *TcpProtocol) Close() error {
	if p.Conn != nil || !reflect.ValueOf(p.Conn).IsNil() {
		return p.Conn.Close()
	}
	return nil
}

func (p *TcpProtocol) SendPacket(pk *Packet) error {
	// FIXME length, header position wrong.
	// log.Println("Sending:", pk.ClientId, pk.Header, pk.MsgBuff)
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
	// log.Println("Write header", pk.Header)
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Header); err != nil {
		return err
	}
	// log.Println("Write Length:", pk.Length)

	if err := binary.Write(p.Writer, binary.BigEndian, pk.Length); err != nil {
		return err
	}
	// log.Println("Write Buff:", pk.MsgBuff)
	if _, err := p.Writer.Write(pk.MsgBuff); err != nil {
		return err
	}
	return p.Writer.Flush()
}

func (p *TcpProtocol) ReadPacket() (*Packet, error) {
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
	// log.Println("Read Header", header)
	if err != nil {
		return nil, err
	}

	subType := header.Flag & FlagBitsType

	// read length
	var length TLength
	err = binary.Read(p.Reader, binary.BigEndian, &length)
	// log.Println("Read Length", length)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(p.Reader, buf)
	// log.Println("Read buff", buf)
	if err != nil {
		return nil, err
	}
	// log.Println("return packet", buf)
	packet := &Packet{
		ClientId: clientId,
		SubType:  subType,
		Length:   length,
		Header:   header,
		MsgBuff:  buf,
	}
	return packet, nil
}
