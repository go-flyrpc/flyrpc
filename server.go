package fly

import (
	"log"
	"net"
)

type Server struct {
	Router       Router
	multiplex    bool
	serializer   Serializer
	listener     net.Listener
	transports   []*transport
	contextMap   map[int]*Context
	nextClientId int
}

type transport struct {
	protocol  Protocol
	server    *Server
	clientIds []int
}

func NewServer() *Server {
	return nil
}

func (s *Server) Broadcast(clientIds []int, cmdId CmdIdSize, v Message) error {
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

func (s *Server) OnMessage(cmdId CmdIdSize, handler HandlerFunc) {
}

func (s *Server) SendMessage(clientId CmdIdSize, cmdId CmdIdSize, v Message) error {
	return nil
}

func (s *Server) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	go s.handleConnections()
	return nil
}

func (s *Server) handleConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Fatal("Accept error", err)
		}
		s.transports = append(s.transports, newTransport(conn, s))
	}
}

func newTransport(conn net.Conn, server *Server) *transport {
	protocol := NewProtocol(conn, server.IsMultiplex())
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
	} else {
		transport.addClient(server.GetNextClientId())
	}
	protocol.OnPacket(transport.emitPacket)
	return transport
}

func (t *transport) emitPacket(pkt *Packet) {
	// TODO fix non clientId
	clientId := pkt.ClientId
	t.getContext(clientId).emitPacket(pkt)
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
	delete(t.server.contextMap, clientId)
	return context
}

func (t *transport) Close() error {
	t.clientIds = nil
	// remove all clients
	for id := range t.clientIds {
		delete(t.server.contextMap, id)
	}
	return nil
}
