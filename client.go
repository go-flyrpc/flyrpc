package fly

import "net"

type Client struct {
	*Context
	Router     Router
	serializer Serializer
}

func NewClient(conn net.Conn, serializer Serializer) *Client {
	protocol := NewProtocol(conn, false)
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 0, serializer)
	protocol.OnPacket(context.emitPacket)
	return &Client{
		context,
		router,
		serializer,
	}
}

func (c *Client) OnMessage(cmdId CmdIdSize, handler HandlerFunc) {
	c.Router.AddRoute(cmdId, handler)
}
