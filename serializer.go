package fly

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Message interface{}

type Serializer interface {
	Marshal(Message) ([]byte, error)
	Unmarshal([]byte, Message) error
}

var (
	JSON     Serializer = &_json{}
	Protobuf Serializer = &_proto{}
	Msgpack  Serializer = &_msgpack{}
)

type _json struct {
}

func (j *_json) Marshal(v Message) ([]byte, error) {
	return json.Marshal(v)
}

func (j *_json) Unmarshal(bytes []byte, v Message) error {
	return json.Unmarshal(bytes, v)
}

type _proto struct {
}

func (p *_proto) Marshal(v Message) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, NewFlyError(ErrNotProtoMessage, nil)
	}
	return proto.Marshal(m)
}

func (p *_proto) Unmarshal(bytes []byte, v Message) error {

	m, ok := v.(proto.Message)
	if !ok {
		return NewFlyError(ErrNotProtoMessage, nil)
	}
	return proto.Unmarshal(bytes, m)
}

type _msgpack struct {
}

func (m *_msgpack) Marshal(v Message) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (m *_msgpack) Unmarshal(bytes []byte, v Message) error {
	return msgpack.Unmarshal(bytes, v)
}
