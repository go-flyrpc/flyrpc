package flyrpc

import (
	"log"
	"time"
)

type Context struct {
	Protocol Protocol
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

func (ctx *Context) SendError(cmd string, seq TSeq, err Error) error {
	buff := []byte(err.Code())
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

func (ctx *Context) Call(cmd string, reply Message, message Message) error {
	log.Println(ctx.ClientId, "Call", cmd, message)
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	packet := &Packet{
		Flag:    TypeRPC | RPCFlagReq,
		Cmd:     cmd,
		Seq:     ctx.getNextSeq(),
		MsgBuff: buff,
	}

	// Send Packet
	if err := ctx.Protocol.SendPacket(packet); err != nil {
		return err
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
			return newReplyError(string(rPacket.MsgBuff), rPacket)
		}
		if reply != nil {
			return ctx.serializer.Unmarshal(rPacket.MsgBuff, reply)
		}
		return nil
	case <-time.After(ctx.timeout):
		return NewFlyError(ErrTimeOut)
	}
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
		return NewFlyError(ErrTimeOut)
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
			log.Println(ctx.ClientId, "No channel found, pkt is :", pkt)
			return
		}
		replyChan <- pkt
		return
	}
	ctx.Packet = pkt
	log.Println(ctx.ClientId, "OnMessage", pkt.Cmd)
	if err := ctx.Router.emitPacket(ctx, pkt); err != nil {
		log.Println(ctx.ClientId, "Error to call packet", err)
	}
}

func (ctx *Context) emitPingPacket(pkt *Packet) {
	if pkt.Flag&PingFlagPing != 0 {
		log.Println(ctx.ClientId, "sendPong")
		ctx.sendPong(pkt)
	} else if pkt.Flag&PingFlagPong != 0 {
		log.Println(ctx.ClientId, "recvPong")
		ctx.pingChans[pkt.Seq] <- pkt.MsgBuff
	}
}

func (ctx *Context) getNextSeq() TSeq {
	ctx.nextSeq++
	return ctx.nextSeq
}
