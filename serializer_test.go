package fly

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

type TestNoneProto struct {
	Id   int32
	Name string
}

type TestUser struct {
	Id   int32  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
}

func (m *TestUser) Reset()         { *m = TestUser{} }
func (m *TestUser) String() string { return proto.CompactTextString(m) }
func (*TestUser) ProtoMessage()    {}

func testSerializer(t *testing.T, s Serializer) {
	bytes, err := s.Marshal(&TestUser{Id: 123, Name: "abc"})
	assert.Nil(t, err)
	u := &TestUser{}
	err = s.Unmarshal(bytes, u)
	assert.Nil(t, err)
	assert.Equal(t, 123, u.Id)
	assert.Equal(t, "abc", u.Name)
}

func TestJSONSerializer(t *testing.T) {
	testSerializer(t, JSON)
}

func TestProtoSerializer(t *testing.T) {
	s := Protobuf
	bytes, err := s.Marshal(&TestNoneProto{Id: 123, Name: "abc"})
	assert.NotNil(t, err)
	bytes, err = s.Marshal(&TestUser{Id: 123, Name: "abc"})
	assert.Nil(t, err)
	m := &TestNoneProto{}
	err = s.Unmarshal(bytes, m)
	assert.NotNil(t, err)
	u := &TestUser{}
	err = s.Unmarshal(bytes, u)
	assert.Nil(t, err)
	assert.Equal(t, 123, u.Id)
	assert.Equal(t, "abc", u.Name)
}

func TestMsgpack(t *testing.T) {
	testSerializer(t, Msgpack)
}
