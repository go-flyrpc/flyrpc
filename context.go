package fly

import (
	"log"
	"time"
)

type Context struct {
	Protocol Protocol
	ClientId int
	Data     map[string]interface{}
	Packet   *Packet
	Router   Router
	// private
	serializer Serializer
	nextSeq    TSeq
	pingSeq    TSeq
	replyChans map[int]chan []byte
	pingChans  map[TSeq]chan []byte
	timeout    time.Duration
}

func NewContext(protocol Protocol, router Router, clientId int, serializer Serializer) *Context {
	return &Context{
		Protocol:   protocol,
		Router:     router,
		ClientId:   clientId,
		serializer: serializer,
		replyChans: make(map[int]chan []byte),
		pingChans:  make(map[TSeq]chan []byte),
		timeout:    10 * time.Second,
	}
}

func (ctx *Context) SendPacket(flag byte, cmd TCmd, seq TSeq, buff []byte) error {
	return ctx.Protocol.SendPacket(&Packet{
		ClientId: ctx.ClientId,
		Header: &Header{
			Flag: flag,
			Cmd:  cmd,
			Seq:  seq,
		},
		MsgBuff: buff,
	})
}

func (ctx *Context) SendError(err *flyError) error {
	// TODO
	return nil
}

func (ctx *Context) SendMessage(cmd TCmd, message Message) error {
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	return ctx.SendPacket(TypeRPC, cmd, ctx.getNextSeq(), buff)
}

func (ctx *Context) Call(cmd TCmd, reply Message, message Message) error {
	log.Println(ctx.ClientId, "Call", cmd, message)
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	header := &Header{
		Flag: TypeRPC,
		Cmd:  cmd,
		Seq:  ctx.getNextSeq(),
	}
	if reply == nil {
		panic("reply message can't be nil")
	}

	// Send Packet
	if err := ctx.Protocol.SendPacket(&Packet{Header: header, MsgBuff: buff}); err != nil {
		return err
	}

	// init channel before send packet
	replyChan := make(chan []byte, 1)
	// set replyChan for cmd | seq
	chanId := ctx.getChanId(header)
	ctx.replyChans[chanId] = replyChan

	// make sure that replyChan is released
	defer delete(ctx.replyChans, chanId)
	select {
	case rBuff := <-replyChan:
		return ctx.serializer.Unmarshal(rBuff, reply)
	case <-time.After(ctx.timeout):
		return NewFlyError(ErrTimeOut)
	}
}

func (ctx *Context) Ping(length TLength, timeout time.Duration) error {
	ctx.pingSeq++
	seq := ctx.pingSeq
	err := ctx.Protocol.Ping(seq, length)
	if err != nil {
		return err
	}
	pingChan := make(chan []byte, 1)
	ctx.pingChans[seq] = pingChan
	defer delete(ctx.pingChans, seq)
	select {
	case <-pingChan:
	case <-time.After(timeout):
		return NewFlyError(ErrTimeOut)
	}
	return nil
}

func (ctx *Context) emitPacket(pkt *Packet) {
	subType := pkt.Header.Flag & FlagBitsType
	switch subType {
	case TypeRPC:
		ctx.emitRPCPacket(pkt)
	case TypePing:
		ctx.emitPingPacket(pkt)
	default:
		log.Println("Unsupported subType", subType)
	}
}

func (ctx *Context) emitRPCPacket(pkt *Packet) {
	if pkt.Header.Flag&RPCFlagResp != 0 {
		chanId := ctx.getChanId(pkt.Header)
		replyChan := ctx.replyChans[chanId]
		if replyChan == nil {
			log.Println(ctx.ClientId, "No channel found, pkt is :", pkt.Header, chanId)
			return
		}
		replyChan <- pkt.MsgBuff
		return
	}
	ctx.Packet = pkt
	log.Println(ctx.ClientId, "OnMessage", pkt.Header.Cmd)
	if err := ctx.Router.emitPacket(ctx, pkt); err != nil {
		log.Println(ctx.ClientId, "Error to call packet", err)
	}
}

func (ctx *Context) emitPingPacket(pkt *Packet) {
	if pkt.Header.Flag&PingFlagPing != 0 {
		ctx.Protocol.Pong(pkt)
	} else if pkt.Header.Flag&PingFlagPong != 0 {
		ctx.pingChans[pkt.Header.Seq] <- pkt.MsgBuff
	}
}

func (ctx *Context) getNextSeq() TSeq {
	ctx.nextSeq++
	return ctx.nextSeq
}

func (ctx *Context) getChanId(header *Header) int {
	return int(header.Cmd)<<16 | int(header.Seq)
}
