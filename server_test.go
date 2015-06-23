package flyrpc

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeClient(t *testing.T, addr string) *Client {
	client, err := Dial("tcp", addr)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	client.SetSerializer(Protobuf)
	return client
}

func TestServerMultiConnect(t *testing.T) {
	server := NewServer(&ServerOpts{
		Serializer: Protobuf,
	})
	wg := &sync.WaitGroup{}
	server.OnConnect(func(ctx *Context) {
		ctx.SendMessage("1", &TestUser{Id: 123})
	})
	go func() {
		err := server.Listen("tcp", "127.0.0.1:15555")
		assert.Nil(t, err)
	}()
	<-time.After(10 * time.Millisecond)
	wg.Add(1)
	c1 := makeClient(t, "127.0.0.1:15555")
	assert.NotNil(t, c1)
	c1.OnMessage("1", func(ctx *Context, user *TestUser) {
		assert.Equal(t, int32(123), user.Id)
		wg.Done()
	})
	wg.Add(1)
	c2 := makeClient(t, "127.0.0.1:15555")
	assert.NotNil(t, c2)
	c2.OnMessage("1", func(ctx *Context, user *TestUser) {
		assert.Equal(t, int32(123), user.Id)
		wg.Done()
	})
	wg.Wait()
	server.Close()
}

func TestServerHandlePacket(t *testing.T) {
	uid := int32(123)
	server := NewServer(&ServerOpts{
		Serializer: Protobuf,
	})
	wg := &sync.WaitGroup{}
	server.OnMessage("1", func(ctx *Context, u *TestUser) {
		assert.Equal(t, uid, u.Id)
		wg.Done()
	})
	go func() {
		err := server.Listen("tcp", "127.0.0.1:15556")
		assert.Nil(t, err)
	}()
	<-time.After(10 * time.Millisecond)
	wg.Add(1)
	client := makeClient(t, "127.0.0.1:15556")
	client.SendMessage("1", &TestUser{Id: 123})
	wg.Wait()
	server.Close()
}

/*
func TestServer(t *testing.T) {
	server := NewServer(&ServerOpts{
		Serializer: Protobuf,
	})
	err := server.Listen("127.0.0.1:15555")
	assert.Nil(t, err)
	assert.NotNil(t, server.listener)

	server.OnMessage(21, func(ctx *Context, in *TestUser) *TestUser {
		reply := new(TestUser)
		log.Println("server on", in)
		// call client cmd 2 and response another user
		log.Println("server call 2")
		err := ctx.Call(12, in, reply)
		assert.Nil(t, err)
		log.Println("client response", reply)
		reply.Id += 10
		return reply
	})

	server.OnConnect(func(ctx *Context) {
		log.Println("connected")
		in := new(TestUser)
		err := ctx.SendMessage(11, in)
		assert.Nil(t, err)
	})

	go server.HandleConnections()

	wg := &sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		testClient(t, wg, i)
	}
	wg.Wait()
}

func testClient(t *testing.T, wg *sync.WaitGroup, i int) {
	wg.Add(1)
	conn, err := net.Dial("tcp", "127.0.0.1:15555")
	assert.Nil(t, err)
	client := NewClient(conn, Protobuf)
	client.Context.ClientId = 100 + i
	client.OnMessage(12, func(ctx *Context, in *TestUser) *TestUser {
		log.Println("client on", in)
		return &TestUser{Id: in.Id + 1}
	})
	reply := new(TestUser)
	client.OnMessage(11, func(ctx *Context, in *TestUser) {
		log.Println("client call 1")
		err = client.Call(21, &TestUser{Id: 100}, reply)
		assert.Nil(t, err)
		log.Println("server response", reply)
		assert.Equal(t, 111, reply.Id)
		wg.Done()
	})
}
*/
