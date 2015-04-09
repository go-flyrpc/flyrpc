package fly

import "net"

// Client use to connect server.
type Client struct {
	// extend with *Context
	*Context
	Router     Router
	serializer Serializer
}

// Create new Client instance.
func NewClient(conn net.Conn, serializer Serializer) *Client {
	protocol := NewProtocol(conn, false)
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 99, serializer)
	protocol.OnPacket(context.emitPacket)
	return &Client{
		context,
		router,
		serializer,
	}
}

func (c *Client) OnMessage(cmd TCmd, handler HandlerFunc) {
	c.Router.AddRoute(cmd, handler)
}
