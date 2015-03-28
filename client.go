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
	return &Client{
		NewContext(protocol, router, 0, serializer),
		router,
		serializer,
	}
}

func (c *Client) OnMessage(cmdId CmdIdSize, handler HandlerFunc) {
	c.Router.AddRoute(cmdId, handler)
}
