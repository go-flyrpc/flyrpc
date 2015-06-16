package flyrpc

import (
	"bufio"
	"bytes"
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
	// TODO zip
	// TODO crc
	// buff = zip(buff)
	// }

	// flag + trans-flag + seq + \n = 5 byte
	pk.Length = TLength(5 + len(pk.MsgBuff) + len(pk.Cmd))
	if pk.Length > MaxLength {
		return NewFlyError(ErrBuffTooLong, nil)
	}

	// write Length
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Length); err != nil {
		return err
	}

	// write Flag
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Flag); err != nil {
		return err
	}

	// write Transfer Flag
	if err := binary.Write(p.Writer, binary.BigEndian, pk.TransferFlag); err != nil {
		return err
	}

	// write Seq
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Seq); err != nil {
		return err
	}

	// write Cmd
	if _, err := p.Writer.WriteString(pk.Cmd + "\n"); err != nil {
		return err
	}

	// write Buff
	if _, err := p.Writer.Write(pk.MsgBuff); err != nil {
		return err
	}
	return p.Writer.Flush()
}

func (p *TcpProtocol) ReadPacket() (*Packet, error) {
	pkt := &Packet{}
	// only for server
	if p.IsMultiplex {
		err := binary.Read(p.Reader, binary.BigEndian, &pkt.ClientId)
		if err != nil {
			return nil, err
		}
	}
	// read length
	err := binary.Read(p.Reader, binary.BigEndian, &pkt.Length)
	if err != nil {
		return nil, err
	}

	// read Full Packet
	buf := make([]byte, pkt.Length)
	_, err = io.ReadFull(p.Reader, buf)
	if err != nil {
		return nil, err
	}
	// TODO checksum
	// TODO unzip

	reader := bytes.NewBuffer(buf)

	// read Flag
	pkt.Flag, err = reader.ReadByte()
	if err != nil {
		return nil, err
	}
	pkt.SubType = pkt.Flag & FlagBitsType

	// read TransferFlag
	pkt.TransferFlag, err = reader.ReadByte()
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

	// read Cmd
	cmd, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	pkt.Cmd = cmd[:len(cmd)-1]

	// read MsgBuff
	pkt.MsgBuff = reader.Bytes()
	return pkt, nil
}
