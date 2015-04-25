package flyrpc

// TypeBits - bits of sub protocol
// TypeRPC  - type of RPC. Main feature
// TypePing - type of Ping. Keepalive
// TypeHello - type of Hello. Tell the client information related with protocol, like version, zip, supported encoding
const (
	FlagBitsType byte = 0xb0 // 11000000
	TypeRPC      byte = 0xb0 // 11000000
	TypePing     byte = 0x80 // 10000000
	TypeHello    byte = 0x40 // 01000000
	TypeMQ       byte = 0x00 // 00000000
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
const MaxCommand = ^TCmd(0)

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
