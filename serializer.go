package flyrpc

import (
	"encoding/json"
	"reflect"
)

type Message interface{}

var (
	typeBytes  = reflect.TypeOf([]byte{})
	typeString = reflect.TypeOf("")
)

func MessageToBytes(message Message, serializer Serializer) ([]byte, error) {
	messageType := reflect.TypeOf(message)
	if messageType == typeBytes {
		return message.([]byte), nil
	}
	if messageType == typeString {
		return []byte(message.(string)), nil
	}
	return serializer.Marshal(message)
}

type Serializer interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

type serializer struct {
	marshal   func(interface{}) ([]byte, error)
	unmarshal func([]byte, interface{}) error
}

func NewSerializer(marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) Serializer {
	return &serializer{
		marshal:   marshal,
		unmarshal: unmarshal,
	}
}

func (s *serializer) Marshal(msg interface{}) ([]byte, error) {
	return s.marshal(msg)
}

func (s *serializer) Unmarshal(bytes []byte, msg interface{}) error {
	return s.unmarshal(bytes, msg)
}

var (
	JSON Serializer = &serializer{marshal: json.Marshal, unmarshal: json.Unmarshal}
	/*
		Protobuf Serializer = &_proto{}
		Msgpack  Serializer = &_msgpack{}
	*/
)

/*
type _proto struct {
}

func (p *_proto) Marshal(v interface{}) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, newError(ErrNotProtoMessage)
	}
	return proto.Marshal(m)
}

func (p *_proto) Unmarshal(bytes []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if !ok {
		return newError(ErrNotProtoMessage)
	}
	return proto.Unmarshal(bytes, m)
}

type _msgpack struct {
}

func (m *_msgpack) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (m *_msgpack) Unmarshal(bytes []byte, v interface{}) error {
	return msgpack.Unmarshal(bytes, v)
}
*/
