package flyrpc

import (
	"io"
	"log"
	"net"
)

type ServerOpts struct {
	Serializer Serializer
	Multiplex  bool
}

type Server struct {
	Router          Router
	multiplex       bool
	serializer      Serializer
	listener        net.Listener
	transports      []*transport
	contextMap      map[int]*Context
	connectHandlers []func(*Context)
	nextClientId    int
}

type transport struct {
	protocol  Protocol
	server    *Server
	multiplex bool
	context   *Context
	clientIds []int
}

func NewServer(opts *ServerOpts) *Server {
	if opts.Serializer == nil {
		opts.Serializer = JSON
	}
	return &Server{
		Router:          NewRouter(opts.Serializer),
		multiplex:       opts.Multiplex,
		serializer:      opts.Serializer,
		transports:      make([]*transport, 0),
		contextMap:      make(map[int]*Context),
		connectHandlers: make([]func(*Context), 0),
		nextClientId:    0,
	}
}

func (s *Server) Broadcast(clientIds []int, cmd string, v Message) error {
	return nil
}

func (s *Server) GetContext(clientId int) *Context {
	// TODO 考虑多路复用情况, 多个client会共享一个transport
	return s.contextMap[clientId]
}

func (s *Server) GetNextClientId() int {
	s.nextClientId++
	return s.nextClientId
}

func (s *Server) IsMultiplex() bool {
	return s.multiplex
}

func (s *Server) OnConnect(connectHandler func(*Context)) {
	s.connectHandlers = append(s.connectHandlers, connectHandler)
}

func (s *Server) OnMessage(cmd string, handler HandlerFunc) {
	s.Router.AddRoute(cmd, handler)
}

func (s *Server) emitContext(ctx *Context) {
	for _, handler := range s.connectHandlers {
		go handler(ctx)
	}
}

func (s *Server) SendMessage(clientId int, cmd string, v Message) error {
	return s.GetContext(clientId).SendMessage(cmd, v)
}

func (s *Server) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.handleConnections()
	return nil
}

func (s *Server) Close() error {
	for _, t := range s.transports {
		t.Close()
	}
	err := s.listener.Close()
	return err
}

func (s *Server) handleConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Accept error", err)
			break
		} else {
			log.Println("New Connection", conn.RemoteAddr())
		}
		s.transports = append(s.transports, newTransport(conn, s))
	}
}

func newTransport(conn net.Conn, server *Server) *transport {
	protocol := NewTcpProtocol(conn, server.IsMultiplex())
	transport := &transport{
		protocol: protocol,
		server:   server,
	}
	if server.IsMultiplex() {
		// DO NOTHING
		// TODO somewhere wait message to add clientId
		// For a frontend multiplex server
		// TODO Make standalone frontend server
		// TODO dispatch clientId to a connected backend server
		transport.multiplex = true
	} else {
		ctx := transport.addClient(server.GetNextClientId())
		transport.context = ctx
		server.emitContext(ctx)
	}
	go transport.handlePackets()
	return transport
}

func (t *transport) handlePackets() {
	for {
		packet, err := t.protocol.ReadPacket()
		if err != nil {
			if err != io.EOF {
				log.Println("Close on error", err)
			}
			t.Close()
			break
		}
		go t.emitPacket(packet)
	}
}

func (t *transport) emitPacket(pkt *Packet) {
	if t.multiplex {
		clientId := pkt.ClientId
		t.getContext(clientId).emitPacket(pkt)
	} else {
		t.context.emitPacket(pkt)
	}
}

func (t *transport) getContext(clientId int) *Context {
	context := t.server.contextMap[clientId]
	if context == nil {
		return t.addClient(clientId)
	}
	return context
}

func (t *transport) addClient(clientId int) *Context {
	t.clientIds = append(t.clientIds, clientId)
	context := NewContext(t.protocol, t.server.Router, clientId, t.server.serializer)
	t.server.contextMap[clientId] = context
	return context
}

func (t *transport) removeClient(clientId int) *Context {
	// TODO remove clientId from clientIds
	// remove context from server.contextMap
	context := t.server.contextMap[clientId]
	if context != nil {
		context.Close()
	}
	delete(t.server.contextMap, clientId)
	return context
}

func (t *transport) Close() error {
	// remove all clients
	for id := range t.clientIds {
		t.removeClient(id)
	}
	t.clientIds = nil
	return t.protocol.Close()
}
