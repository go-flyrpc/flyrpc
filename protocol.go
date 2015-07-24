package flyrpc

// TypeBits - bits of sub protocol
// TypeRPC  - type of RPC. Main feature
// TypePing - type of Ping. Keepalive
// TypeHello - type of Hello. Tell the client information related with protocol, like version, zip, supported encoding
const (
	FlagResponse     byte = 0x80
	FlagWaitResponse byte = 0x40
	FlagZipCode      byte = 0x08
	FlagZipPayload   byte = 0x04
	FlagLenPayload   byte = 0x03
)

type TSeq uint16
type TLength uint64

const MaxLength = ^TLength(0)

type Packet struct {
	// TODO remove this from packet
	Protocol Protocol
	ClientId int

	Flag byte
	// message sequence
	Seq     TSeq
	Length  TLength
	Code    string
	Payload []byte
}

type Protocol interface {
	ReadPacket() (*Packet, error)
	SendPacket(*Packet) error
	Close() error
}
