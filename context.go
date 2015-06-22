package flyrpc

import (
	"log"
	"time"
)

type Context struct {
	Protocol Protocol
	Tag      string
	ClientId int
	Session  interface{}
	Packet   *Packet
	Router   Router
	// private
	serializer Serializer
	nextSeq    TSeq
	pingSeq    TSeq
	replyChans map[TSeq]chan *Packet
	pingChans  map[TSeq]chan []byte
	timeout    time.Duration
	// close handler
	closeHandler func(*Context)
}

func NewContext(protocol Protocol, router Router, clientId int, serializer Serializer) *Context {
	return &Context{
		Protocol:   protocol,
		Router:     router,
		ClientId:   clientId,
		serializer: serializer,
		replyChans: make(map[TSeq]chan *Packet),
		pingChans:  make(map[TSeq]chan []byte),
		timeout:    10 * time.Second,
	}
}

func (ctx *Context) SendPacket(flag byte, cmd string, seq TSeq, buff []byte) error {
	return ctx.Protocol.SendPacket(&Packet{
		ClientId: ctx.ClientId,
		Flag:     flag,
		Cmd:      cmd,
		Seq:      seq,
		MsgBuff:  buff,
	})
}

func (ctx *Context) SendError(cmd string, seq TSeq, err error) error {
	buff := []byte(err.Error())
	return ctx.SendPacket(
		TypeRPC|RPCFlagResp|RPCFlagError,
		cmd,
		seq,
		buff,
	)
}

func (ctx *Context) SendMessage(cmd string, message Message) error {
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	return ctx.SendPacket(TypeRPC, cmd, ctx.getNextSeq(), buff)
}

func (ctx *Context) GetReply(cmd string, message Message) ([]byte, error) {
	log.Println(ctx.Tag, "Call", cmd, message)

	buff, err := MessageToBytes(message, ctx.serializer)
	if err != nil {
		return nil, err
	}
	packet := &Packet{
		Flag:    TypeRPC | RPCFlagReq,
		Cmd:     cmd,
		Seq:     ctx.getNextSeq(),
		MsgBuff: buff,
	}

	// Send Packet
	if err := ctx.Protocol.SendPacket(packet); err != nil {
		return nil, err
	}

	// init channel before send packet
	replyChan := make(chan *Packet, 1)
	// set replyChan for cmd | seq
	ctx.replyChans[packet.Seq] = replyChan

	// make sure that replyChan is released
	defer delete(ctx.replyChans, packet.Seq)
	select {
	case rPacket := <-replyChan:
		log.Println("reply buff", rPacket.MsgBuff)
		if rPacket.Flag&RPCFlagError != 0 {
			log.Println("reply error", string(rPacket.MsgBuff))
			return nil, newReplyError(string(rPacket.MsgBuff), rPacket)
		}
		return rPacket.MsgBuff, nil

	case <-time.After(ctx.timeout):
		return nil, newError(ErrTimeOut)
	}
}

func (ctx *Context) Call(cmd string, reply Message, message Message) error {
	bytes, err := ctx.GetReply(cmd, message)
	if err != nil {
		return err
	}
	if reply != nil {
		ctx.serializer.Unmarshal(bytes, reply)
	}
	return nil
}

func (ctx *Context) sendPingPacket(pingFlag byte, seq TSeq, bytes []byte) error {
	return ctx.Protocol.SendPacket(&Packet{
		Flag:    TypePing | pingFlag,
		Cmd:     "",
		Seq:     seq,
		Length:  TLength(len(bytes)),
		MsgBuff: bytes,
	})
}

func (ctx *Context) sendPing(seq TSeq, length TLength) error {
	return ctx.sendPingPacket(PingFlagPing, seq, make([]byte, length))
}

func (ctx *Context) sendPong(pkt *Packet) error {
	return ctx.sendPingPacket(PingFlagPong, pkt.Seq, pkt.MsgBuff)
}

func (ctx *Context) Ping(length TLength, timeout time.Duration) error {
	ctx.pingSeq++
	seq := ctx.pingSeq
	err := ctx.sendPing(seq, length)
	if err != nil {
		return err
	}
	pingChan := make(chan []byte, 1)
	ctx.pingChans[seq] = pingChan
	defer delete(ctx.pingChans, seq)
	select {
	case <-pingChan:
	case <-time.After(timeout):
		return newError(ErrTimeOut)
	}
	return nil
}

func (ctx *Context) emitPacket(pkt *Packet) {
	subType := pkt.Flag & FlagBitsType
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
	if pkt.Flag&RPCFlagResp != 0 {
		replyChan := ctx.replyChans[pkt.Seq]
		if replyChan == nil {
			log.Println(ctx.Tag, "No channel found, pkt is :", pkt)
			return
		}
		replyChan <- pkt
		return
	}
	ctx.Packet = pkt
	log.Println(ctx.Tag, "OnMessage", pkt.Cmd, pkt.Flag, pkt.MsgBuff)
	if err := ctx.Router.emitPacket(ctx, pkt); err != nil {
		log.Println(ctx.Tag, "Error to call packet", err)
	}
}

func (ctx *Context) emitPingPacket(pkt *Packet) {
	if pkt.Flag&PingFlagPing != 0 {
		log.Println(ctx.Tag, "sendPong")
		ctx.sendPong(pkt)
	} else if pkt.Flag&PingFlagPong != 0 {
		log.Println(ctx.Tag, "recvPong")
		ctx.pingChans[pkt.Seq] <- pkt.MsgBuff
	}
}

func (ctx *Context) getNextSeq() TSeq {
	ctx.nextSeq++
	return ctx.nextSeq
}

func (ctx *Context) OnClose(handler func(*Context)) {
	ctx.closeHandler = handler
}

func (ctx *Context) Close() {
	if ctx.closeHandler != nil {
		ctx.closeHandler(ctx)
	}
}
