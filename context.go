package fly

import "log"

type Context struct {
	Protocol Protocol
	ClientId int
	Data     map[string]interface{}
	Packet   *Packet
	Router   Router
	// private
	serializer Serializer
	nextSeq    TSeq
	replyChans map[int]chan<- []byte
}

func NewContext(protocol Protocol, router Router, clientId int, serializer Serializer) *Context {
	return &Context{
		Protocol:   protocol,
		Router:     router,
		ClientId:   clientId,
		serializer: serializer,
		replyChans: make(map[int]chan<- []byte),
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
	return ctx.SendPacket(0, cmd, ctx.getNextSeq(), buff)
}

func (ctx *Context) Call(cmd TCmd, reply Message, message Message) error {
	log.Println("Call", cmd, message)
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	header := &Header{
		Cmd: cmd,
		Seq: ctx.getNextSeq(),
	}
	if reply == nil {
		panic("reply message can't be nil")
	}
	// init channel before send packet
	replyChan := make(chan []byte, 1)
	// set replyChan for cmd | seq
	chanId := ctx.getChanId(header)
	ctx.replyChans[chanId] = replyChan
	log.Println("make chan", ctx, chanId, ctx.replyChans)
	defer delete(ctx.replyChans, chanId)
	log.Println("make chan", ctx, chanId, ctx.replyChans)
	if err := ctx.Protocol.SendPacket(&Packet{Header: header, MsgBuff: buff}); err != nil {
		return err
	}
	log.Println("make chan", ctx, chanId, ctx.replyChans)
	// wait to get response
	rBuff := <-replyChan
	log.Println("make chan", ctx, chanId, ctx.replyChans)
	return ctx.serializer.Unmarshal(rBuff, reply)
}

func (ctx *Context) emitPacket(pkt *Packet) {
	if pkt.Header.Flag&LFlagResp != 0 {
		chanId := ctx.getChanId(pkt.Header)
		replyChan := ctx.replyChans[chanId]
		if replyChan == nil {
			log.Println("No channel found, pkt is :", pkt.Header, chanId)
			log.Println("replyChans", ctx, ctx.replyChans)
			return
		}
		replyChan <- pkt.MsgBuff
		return
	}
	ctx.Packet = pkt
	if err := ctx.Router.emitPacket(ctx, pkt); err != nil {
		log.Println("Error to call packet", err)
	}
}

func (ctx *Context) getNextSeq() TSeq {
	ctx.nextSeq++
	return ctx.nextSeq
}

func (ctx *Context) getChanId(header *Header) int {
	return int(header.Cmd)<<16 | int(header.Seq)
}
