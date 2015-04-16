package fly

import (
	"io"
	"log"
	"net"
)

// Client use to connect server.
type Client struct {
	// extend with *Context
	*Context
	Protocol   Protocol
	Router     Router
	serializer Serializer
}

// Create new Client instance.
func NewClient(conn net.Conn, serializer Serializer) *Client {
	protocol := NewProtocol(conn, false)
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 99, serializer)
	cli := &Client{
		context,
		protocol,
		router,
		serializer,
	}
	go cli.handlePackets()
	return cli
}

func (c *Client) handlePackets() {
	for {
		packet, err := c.Protocol.ReadPacket()
		if err != nil {
			if err != io.EOF {
				log.Println("Close on error", err)
			}
			c.Close()
			break
		}
		c.Context.emitPacket(packet)
	}
}

func (c *Client) OnMessage(cmd TCmd, handler HandlerFunc) {
	c.Router.AddRoute(cmd, handler)
}

func (c *Client) Close() error {
	return c.Protocol.Close()
}
