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
	// if support zip {
	// TODO zip
	// payload = zip(payload)
	// }
	if pk.Length == 0 {
		pk.Length = TLength(len(pk.Payload))
	}
	// write Header
	if err := p.SendHeader(pk); err != nil {
		return err
	}
	if int(pk.Length)+len(pk.Code) > 1300 {
		// flush header first
		if err := p.Writer.Flush(); err != nil {
			return err
		}
	}
	// TODO make Big payload write more effecient
	// write Payload
	if _, err := p.Writer.Write(pk.Payload); err != nil {
		return err
	}
	return p.Writer.Flush()
}

func (p *TcpProtocol) SendHeader(pk *Packet) error {
	var sizeOfLength byte
	if pk.Length > 0xffffffff {
		sizeOfLength = 8
		pk.Flag = pk.Flag | 0x03
	} else if pk.Length > 0xffff {
		sizeOfLength = 4
		pk.Flag = pk.Flag | 0x02
	} else if pk.Length > 0xff {
		sizeOfLength = 2
		pk.Flag = pk.Flag | 0x01
	} else {
		sizeOfLength = 1
	}

	// write Flag
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Flag); err != nil {
		return err
	}

	// write Seq
	if err := binary.Write(p.Writer, binary.BigEndian, pk.Seq); err != nil {
		return err
	}

	// write Code
	if _, err := p.Writer.WriteString(pk.Code); err != nil {
		return err
	}
	if err := p.Writer.WriteByte(0); err != nil {
		return err
	}

	// write Payload Length
	if sizeOfLength == 1 {
		if err := p.Writer.WriteByte(byte(pk.Length)); err != nil {
			return err
		}
	} else if sizeOfLength == 2 {
		if err := binary.Write(p.Writer, binary.BigEndian, uint16(pk.Length)); err != nil {
			return err
		}
	} else if sizeOfLength == 4 {
		if err := binary.Write(p.Writer, binary.BigEndian, uint32(pk.Length)); err != nil {
			return err
		}
	} else if sizeOfLength == 8 {
		if err := binary.Write(p.Writer, binary.BigEndian, uint64(pk.Length)); err != nil {
			return err
		}
	}
	return nil
}

func (p *TcpProtocol) ReadPacket() (*Packet, error) {
	pkt := &Packet{}

	reader := p.Reader

	if err := p.ReadHeader(pkt); err != nil {
		return nil, err
	}

	// read Payload
	pkt.Payload = make([]byte, pkt.Length)
	if _, err := io.ReadFull(reader, pkt.Payload); err != nil {
		return nil, err
	}
	// TODO unzip
	return pkt, nil
}

func (p *TcpProtocol) ReadHeader(pkt *Packet) error {
	reader := p.Reader

	var err error
	// read Flag
	pkt.Flag, err = reader.ReadByte()
	if err != nil {
		return err
	}
	powOfLength := pkt.Flag & FlagLenPayload

	// read Seq
	var seq uint16
	err = binary.Read(reader, binary.BigEndian, &seq)
	if err != nil {
		return err
	}
	pkt.Seq = TSeq(seq)

	// read Code
	code, err := reader.ReadString(0)
	if err != nil {
		return err
	}
	pkt.Code = code[:len(code)-1]

	// read length
	if powOfLength == 0 {
		l, err := reader.ReadByte()
		if err != nil {
			return err
		}
		pkt.Length = TLength(l)
	} else if powOfLength == 1 {
		var l uint16
		err = binary.Read(reader, binary.BigEndian, &l)
		if err != nil {
			return err
		}
		pkt.Length = TLength(l)
	} else if powOfLength == 2 {
		var l uint32
		err = binary.Read(reader, binary.BigEndian, &l)
		if err != nil {
			return err
		}
		pkt.Length = TLength(l)
	} else if powOfLength == 3 {
		var l uint64
		err = binary.Read(reader, binary.BigEndian, &l)
		if err != nil {
			return err
		}
		pkt.Length = TLength(l)
	}
	return nil
}
