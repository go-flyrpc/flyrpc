package flyrpc

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"reflect"
)

type TcpProtocol struct {
	// Conn
	Conn net.Conn
	// Reader
	Reader *bufio.Reader
	// Writer
	Writer *bufio.Writer
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
		Reader: bufio.NewReader(reader),
		Writer: bufio.NewWriter(writer),
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
	// log.Println("Sending:", pk.ClientId, pk.Header, pk.MsgBuff)
	if p.Writer == nil {
		err := p.Close()
		return newFlyError(ErrWriterClosed, err)
	}
	if p.Writer.Available() == 0 {
		return newError(ErrWriterClosed)
	}
	if len(pk.Payload) > int(MaxLength) {
		return newError(ErrBuffTooLong)
	}
	cmdSize := len([]byte(pk.Code))
	if cmdSize > 255 {
		return newError("COMMAND_TOO_LONG")
	}
	// if support zip {
	// TODO zip
	// payload = zip(payload)
	// }

	// write Flag
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Flag); err != nil {
		return err
	}

	// write Seq
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Seq); err != nil {
		return err
	}

	// write CmdSize
	cmdSizeByte := byte(cmdSize)
	if err := binary.Write(p.Writer, binary.BigEndian, cmdSizeByte); err != nil {
		return err
	}

	// write Code
	if _, err := p.Writer.WriteString(pk.Code); err != nil {
		return err
	}

	// write Payload Length
	pk.Length = TLength(len(pk.Payload))
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Length); err != nil {
		return err
	}

	// write Payload
	if _, err := p.Writer.Write(pk.Payload); err != nil {
		return err
	}
	return p.Writer.Flush()
}

func (p *TcpProtocol) ReadPacket() (*Packet, error) {
	pkt := &Packet{}

	reader := p.Reader

	var err error
	// read Flag
	pkt.Flag, err = reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// read Seq
	var seq uint16
	err = binary.Read(reader, binary.BigEndian, &seq)
	if err != nil {
		return nil, err
	}
	pkt.Seq = TSeq(seq)

	// read CmdSize
	cmdSize, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// read Cmd
	cmdBuff := make([]byte, cmdSize)
	_, err = io.ReadFull(reader, cmdBuff)
	if err != nil {
		return nil, err
	}
	pkt.Code = string(cmdBuff)

	// read length
	err = binary.Read(reader, binary.BigEndian, &pkt.Length)
	if err != nil {
		return nil, err
	}

	// read Payload
	pkt.Payload = make([]byte, pkt.Length)
	_, err = io.ReadFull(reader, pkt.Payload)
	if err != nil {
		return nil, err
	}
	// TODO unzip
	return pkt, nil
}
