package flyrpc

import (
	"log"
	"time"
)

type Context struct {
	Protocol Protocol
	Debug    bool
	Tag      string
	ClientId int
	Session  interface{}
	Packet   *Packet
	Router   Router
	// private
	serializer Serializer
	nextSeq    TSeq
	replyChans map[TSeq]chan *Packet
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
		timeout:    10 * time.Second,
	}
}

func (ctx *Context) debug(args ...interface{}) {
	if ctx.Debug {
		if ctx.Tag != "" {
			args = append([]interface{}{"[" + ctx.Tag + "]"}, args...)
		}
		log.Println(args...)
	}
}

func (ctx *Context) sendPacket(flag byte, code string, seq TSeq, payload []byte) error {
	return ctx.Protocol.SendPacket(&Packet{
		ClientId: ctx.ClientId,
		Flag:     flag,
		Code:     code,
		Seq:      seq,
		Payload:  payload,
	})
}

func (ctx *Context) sendError(code string, seq TSeq, err error) error {
	payload := []byte{}
	return ctx.sendPacket(
		FlagResponse,
		err.Error(),
		seq,
		payload,
	)
}

func (ctx *Context) SendMessage(code string, message Message) error {
	payload, err := MessageToBytes(message, ctx.serializer)
	if err != nil {
		return err
	}
	return ctx.sendPacket(FlagWaitResponse, code, ctx.getNextSeq(), payload)
}

func (ctx *Context) GetReply(code string, message Message) ([]byte, error) {
	ctx.debug("Call", code, message)

	payload, err := MessageToBytes(message, ctx.serializer)
	if err != nil {
		return nil, err
	}
	packet := &Packet{
		Flag:    FlagWaitResponse,
		Code:    code,
		Seq:     ctx.getNextSeq(),
		Payload: payload,
	}

	// Send Packet
	if err := ctx.Protocol.SendPacket(packet); err != nil {
		return nil, err
	}

	// init channel before send packet
	replyChan := make(chan *Packet, 1)
	// set replyChan for code | seq
	ctx.replyChans[packet.Seq] = replyChan

	// make sure that replyChan is released
	defer delete(ctx.replyChans, packet.Seq)
	select {
	case rPacket := <-replyChan:
		ctx.debug("reply payload", rPacket.Payload)
		if rPacket.Code != "" {
			ctx.debug("reply error", string(rPacket.Code))
			return nil, newReplyError(string(rPacket.Code), rPacket)
		}
		return rPacket.Payload, nil

	case <-time.After(ctx.timeout):
		return nil, newError(ErrTimeOut)
	}
}

func (ctx *Context) Call(code string, message Message, reply Message) error {
	bytes, err := ctx.GetReply(code, message)
	if err != nil {
		return err
	}
	if reply != nil {
		return ctx.serializer.Unmarshal(bytes, reply)
	}
	return nil
}

func (ctx *Context) GetAsync(code string, message Message) (chan<- []byte, chan<- error) {
	buffChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		bytes, err := ctx.GetReply(code, message)
		buffChan <- bytes
		errChan <- err
	}()
	return buffChan, errChan
}

func (ctx *Context) emitPacket(pkt *Packet) {
	if pkt.Flag&FlagResponse != 0 {
		replyChan := ctx.replyChans[pkt.Seq]
		if replyChan == nil {
			ctx.debug("No channel found, pkt is :", pkt)
			return
		}
		replyChan <- pkt
		return
	}
	ctx.Packet = pkt
	ctx.debug("OnMessage", pkt.Code, pkt.Flag, pkt.Payload)
	if err := ctx.Router.emitPacket(ctx, pkt); err != nil {
		ctx.debug("Error to call packet", err)
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
	ctx.debug("closing")
	ctx.Protocol.Close()
	if ctx.closeHandler != nil {
		ctx.closeHandler(ctx)
	}
}
