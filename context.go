package fly

type Context struct {
	Protocol Protocol
	ClientId int
	Data     map[string]interface{}
	Packet   *Packet
	In       Message
	Out      Message
	Router   Router
}

func NewContext(protocol Protocol, router Router, clientId int) *Context {
	return &Context{Protocol: protocol, Router: router, ClientId: clientId}
}

func (ctx *Context) SendMessage(cmdId CmdIdSize, message Message) error {
	buff, err := ctx.serializer.Marshal(message)
	if err != nil {
		return err
	}
	ctx.Protocol.SendPacket(&Packet{
		ClientId: ctx.ClientId,
		Header: &Header{
			Flag: 0,
		},
		MsgBuff: buff,
	})
	return nil
}

func (ctx *Context) Call() {
}

func (ctx *Context) onPacket(pkt *Packet) {
	ctx.Packet = pkt
	ctx.Router.onPacket(ctx, pkt)
}
