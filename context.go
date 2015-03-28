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
	nextMsgId  MsgIdSize
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

func (ctx *Context) SendPacket(flag byte, cmdId CmdIdSize, msgId MsgIdSize, buff []byte) error {
	return ctx.Protocol.SendPacket(&Packet{
		ClientId: ctx.ClientId,
		Header: &Header{
			Flag:  flag,
			CmdId: cmdId,
			MsgId: msgId,
		},
		MsgBuff: buff,
	})
}

func (ctx *Context) SendError(err *FlyError) error {
	// TODO
	return nil
}

func (ctx *Context) SendMessage(cmdId CmdIdSize, message Message) error {
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	return ctx.SendPacket(0, cmdId, ctx.getNextId(), buff)
}

func (ctx *Context) Call(cmdId CmdIdSize, reply Message, message Message) error {
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	header := &Header{
		CmdId: cmdId,
		MsgId: ctx.getNextId(),
	}
	if reply == nil {
		panic("reply message can't be nil")
	}
	// init channel before send packet
	replyChan := make(chan []byte, 1)
	// set replyChan for cmdId | msgId
	chanId := ctx.getChanId(header)
	ctx.replyChans[chanId] = replyChan
	ctx.Protocol.SendPacket(&Packet{Header: header, MsgBuff: buff})
	// wait to get response
	rBuff := <-replyChan
	delete(ctx.replyChans, chanId)
	return ctx.serializer.Unmarshal(rBuff, reply)
}

func (ctx *Context) emitPacket(pkt *Packet) {
	if pkt.Header.Flag&LFLAG_RESP != 0 {
		chanId := ctx.getChanId(pkt.Header)
		replyChan := ctx.replyChans[chanId]
		if replyChan == nil {
			log.Fatal("No channel found, pkt is :", pkt)
			return
		}
		replyChan <- pkt.MsgBuff
		return
	}
	ctx.Packet = pkt
	ctx.Router.emitPacket(ctx, pkt)
}

func (ctx *Context) getNextId() MsgIdSize {
	ctx.nextMsgId++
	return ctx.nextMsgId
}

func (ctx *Context) getChanId(header *Header) int {
	return int(header.CmdId)<<16 | int(header.MsgId)
}
