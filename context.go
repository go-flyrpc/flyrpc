package flyrpc

import (
	"encoding/binary"
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
	replyChans map[int]chan []byte
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

func (ctx *Context) SendError(cmd TCmd, seq TSeq, err Error) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(err.Code()))
	return ctx.SendPacket(
		TypeRPC|RPCFlagResp|RPCFlagError,
		cmd,
		seq,
		buff,
	)
}

func (ctx *Context) SendMessage(cmd TCmd, message Message) error {
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	return ctx.SendPacket(TypeRPC, cmd, ctx.getNextSeq(), buff)
}

func (ctx *Context) Call(cmd TCmd, reply Message, message Message) error {
	log.Println(ctx.Tag, "Call", cmd, message)
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	header := &Header{
		Flag: TypeRPC | RPCFlagReq,
		Cmd:  cmd,
		Seq:  ctx.getNextSeq(),
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
		if reply != nil {
			return ctx.serializer.Unmarshal(rBuff, reply)
		}
		return nil
	case <-time.After(ctx.timeout):
		return NewFlyError(ErrTimeOut)
	}
}

func (ctx *Context) sendPingPacket(pingFlag byte, seq TSeq, bytes []byte) error {
	return ctx.Protocol.SendPacket(&Packet{
		Header: &Header{
			Flag: TypePing | pingFlag,
			Cmd:  0,
			Seq:  seq,
		},
		Length:  TLength(len(bytes)),
		MsgBuff: bytes,
	})
}

func (ctx *Context) sendPing(seq TSeq, length TLength) error {
	return ctx.sendPingPacket(PingFlagPing, seq, make([]byte, length))
}

func (ctx *Context) sendPong(pkt *Packet) error {
	return ctx.sendPingPacket(PingFlagPong, pkt.Header.Seq, pkt.MsgBuff)
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
	subType := pkt.Header.Flag & FlagBitsType
	switch subType {
	case TypeRPC:
		ctx.emitRPCPacket(pkt)
	case TypePing:
		ctx.emitPingPacket(pkt)
	default:
		log.Println(ctx.Tag, "Unsupported subType", subType)
	}
}

func (ctx *Context) emitRPCPacket(pkt *Packet) {
	if pkt.Header.Flag&RPCFlagResp != 0 {
		chanId := ctx.getChanId(pkt.Header)
		replyChan := ctx.replyChans[chanId]
		if replyChan == nil {
			log.Println(ctx.Tag, "No channel found, pkt is :", pkt.Header, chanId)
			return
		}
		replyChan <- pkt.MsgBuff
		return
	}
	ctx.Packet = pkt
	log.Println(ctx.Tag, "OnMessage", pkt.Header.Cmd)
	if err := ctx.Router.emitPacket(ctx, pkt); err != nil {
		log.Println(ctx.Tag, "Error to call packet", pkt.Header.Cmd, pkt.MsgBuff, err)
	}
}

func (ctx *Context) emitPingPacket(pkt *Packet) {
	if pkt.Header.Flag&PingFlagPing != 0 {
		log.Println(ctx.Tag, "sendPong")
		ctx.sendPong(pkt)
	} else if pkt.Header.Flag&PingFlagPong != 0 {
		log.Println(ctx.Tag, "recvPong")
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

func (ctx *Context) OnClose(handler func(*Context)) {
	ctx.closeHandler = handler
}

func (ctx *Context) Close() {
	if ctx.closeHandler != nil {
		ctx.closeHandler(ctx)
	}
}
