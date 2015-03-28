package fly

import "reflect"

//----------design a ---------
// func(*Packet)
// func(*Packet) Message
// func(*Packet) error
// func(*Packet) (Message, error)
// func(*Packet, Message)
// func(*Packet, Message) Message
// func(*Packet, Message) error
// func(*Packet, Message) (Messag, error)
// ---------- design b -----------
// func(*Packet, in ...Message)
// func(*Packet, in ...Message) Messag
// func(*Packet, in ...Message) error
// func(*Packet, in ...Message) (Messag, error )
//type HandlerFunc interface{}

type HandlerFunc interface{}

type Route interface {
	emitPacket(*Context, *Packet) error
}

type Router interface {
	AddRoute(CmdIdSize, HandlerFunc)
	GetRoute(CmdIdSize) Route
	emitPacket(*Context, *Packet) error
}

type route struct {
	serializer Serializer
	handler    HandlerFunc
	// rpcFunc     RpcFunc
	// messageFunc MessageFunc
	vHandler    reflect.Value
	numIn       int
	numOut      int
	in          []reflect.Type
	out         []reflect.Type
	outErrIndex int
	inType      reflect.Type
	outType     reflect.Type
}

var (
	_err       error
	typeError  reflect.Type = reflect.TypeOf(&_err).Elem()
	typePacket reflect.Type = reflect.TypeOf(&Packet{})
)

func NewRoute(handlerFunc HandlerFunc, s Serializer) *route {
	if s == nil {
		panic("require serializer")
	}
	r := &route{
		serializer:  s,
		handler:     handlerFunc,
		vHandler:    reflect.ValueOf(handlerFunc),
		outErrIndex: -1,
	}
	// FIXME better validate handler
	if r.vHandler.Kind() != reflect.Func {
		panic("handler must be func(...)...")
	}
	numIn := r.vHandler.Type().NumIn()
	numOut := r.vHandler.Type().NumOut()
	r.numIn = numIn
	r.numOut = numOut
	r.in = make([]reflect.Type, numIn)
	r.out = make([]reflect.Type, numOut)
	for i := 0; i < numIn; i++ {
		r.in[i] = r.vHandler.Type().In(i)
	}
	for i := 0; i < numOut; i++ {
		r.out[i] = r.vHandler.Type().Out(i)
	}
	if numIn != 2 {
		panic("Handler must be func(*Packet, Message)")
	}
	if r.in[0] != typePacket {
		panic("handler first parameter must be *Packet")
	}
	r.inType = r.in[1]
	if numOut > 2 {
		panic("Too much returns, handler must return (Message, error) or error")
	}
	if numOut == 2 {
		if !r.out[1].AssignableTo(typeError) {
			panic("Handler should return (Message, error)")
		}
	}
	if numOut > 0 {
		if r.out[numOut-1].AssignableTo(typeError) {
			r.outErrIndex = numOut - 1
		}
		if !r.out[0].AssignableTo(typeError) {
			r.outType = r.out[0]
		}
	}
	return r
}

func (route *route) emitPacket(ctx *Context, pkt *Packet) error {
	v := reflect.New(route.inType.Elem())
	err := route.serializer.Unmarshal(pkt.MsgBuff, v.Interface())
	if err != nil {
		return err
	}
	ret := route.vHandler.Call([]reflect.Value{reflect.ValueOf(pkt), v})
	// retSize := len(ret)
	// if retSize != route.numOut {
	// 	panic("Error result size")
	// }
	// var err error
	if route.outErrIndex >= 0 {
		ve := ret[route.outErrIndex]
		if !ve.IsNil() {
			err = ve.Interface().(error)
		}
	}
	if route.outType != nil {
		// rpc return
		if err != nil {
			flyErr, ok := err.(*FlyError)
			if ok && flyErr.Code < 20000 {
				// client error
				return ctx.SendError(flyErr)
			}
			return err
		}
		vout := ret[0]
		bytes, err := route.serializer.Marshal(vout.Interface())
		if err != nil {
			return err
		}
		return ctx.SendPacket(
			LFLAG_RPC|LFLAG_RESP,
			pkt.Header.CmdId,
			pkt.Header.MsgId,
			bytes)
	}
	return err
}

type router struct {
	routes     map[CmdIdSize]Route
	serializer Serializer
	// routesLock sync.RWMutex
}

func NewRouter(serializer Serializer) Router {
	return &router{routes: make(map[CmdIdSize]Route), serializer: serializer}
}

func (router *router) AddRoute(cmdId CmdIdSize, h HandlerFunc) {
	route := NewRoute(h, router.serializer)
	router.routes[cmdId] = route
}

func (router *router) GetRoute(cmdId CmdIdSize) Route {
	return router.routes[cmdId]
}

func (router *router) emitPacket(ctx *Context, p *Packet) error {
	rt := router.GetRoute(p.Header.CmdId)
	if rt == nil {
		return NewFlyError(ERR_NOT_FOUND, nil)
	}
	return rt.emitPacket(ctx, p)
}
