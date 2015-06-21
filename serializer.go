package flyrpc

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/golang/protobuf/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Message interface{}

var (
	typeBytes  = reflect.TypeOf([]byte{})
	typeString = reflect.TypeOf("")
)

func MessageToBytes(message Message, serializer Serializer) ([]byte, error) {
	messageType := reflect.TypeOf(message)
	log.Println("messageType", messageType)
	if messageType == typeBytes {
		return message.([]byte), nil
	}
	if messageType == typeString {
		return []byte(message.(string)), nil
	}
	return serializer.Marshal(message)
}

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
		return nil, newError(ErrNotProtoMessage)
	}
	return proto.Marshal(m)
}

func (p *_proto) Unmarshal(bytes []byte, v Message) error {
	m, ok := v.(proto.Message)
	if !ok {
		return newError(ErrNotProtoMessage)
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
