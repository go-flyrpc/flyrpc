package flyrpc

import (
	"io"
	"log"
	"net"
)

// Client use to connect server.
type Client struct {
	// extend with *Context
	*Context
}

func Dial(network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	var protocol Protocol
	if network == "tcp" || network == "unix" {
		protocol = NewTcpProtocol(conn, false)
	} else {
		return nil, newError("not support protocol " + network)
	}
	return newClient(protocol, nil), nil
}

func newTcpClient(conn net.Conn, serializer Serializer) *Client {
	protocol := NewTcpProtocol(conn, false)
	return newClient(protocol, serializer)
}

// Create new Client instance.
func newClient(protocol Protocol, serializer Serializer) *Client {
	if serializer == nil {
		serializer = JSON
	}
	router := NewRouter(serializer)
	context := NewContext(protocol, router, 99, serializer)
	cli := &Client{
		context,
	}
	go cli.handlePackets()
	return cli
}

func (c *Client) SetSerializer(serializer Serializer) {
	c.serializer = serializer
	c.Router.(*router).serializer = serializer
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
		go c.emitPacket(packet)
	}
}

func (c *Client) OnMessage(cmd string, handler HandlerFunc) {
	c.Router.AddRoute(cmd, handler)
}

func (c *Client) Close() error {
	return c.Protocol.Close()
}
