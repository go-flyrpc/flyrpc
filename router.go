package fly

import "reflect"

// Message must be explicit type, e.g. *User
// func(*Context)
// func(*Context) Message
// func(*Context) error
// func(*Context) (Message, error)
// func(*Context, Message)
// func(*Context, Message) Message
// func(*Context, Message) error
// func(*Context, Message) (Message, error)

type HandlerFunc interface{}

type Route interface {
	emitPacket(*Context, *Packet) error
}

type Router interface {
	AddRoute(TCmd, HandlerFunc)
	GetRoute(TCmd) Route
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
	_err        error
	typeError   = reflect.TypeOf(&_err).Elem()
	typeContext = reflect.TypeOf(&Context{})
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
	if numIn > 2 {
		panic("Handler arguments must be ([*Context], [Message])")
	}
	if numIn != 2 {
		panic("Handler must be func(*Context, Message)")
	}
	if r.in[0] != typeContext {
		panic("handler first parameter must be *Context")
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
	ret := route.vHandler.Call([]reflect.Value{reflect.ValueOf(ctx), v})
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
			flyErr, ok := err.(*flyError)
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
			LFlagRPC|LFlagResp,
			pkt.Header.Cmd,
			pkt.Header.Seq,
			bytes)
	}
	return err
}

type router struct {
	routes     map[TCmd]Route
	serializer Serializer
	// routesLock sync.RWMutex
}

func NewRouter(serializer Serializer) Router {
	return &router{routes: make(map[TCmd]Route), serializer: serializer}
}

func (router *router) AddRoute(cmd TCmd, h HandlerFunc) {
	route := NewRoute(h, router.serializer)
	router.routes[cmd] = route
}

func (router *router) GetRoute(cmd TCmd) Route {
	return router.routes[cmd]
}

func (router *router) emitPacket(ctx *Context, p *Packet) error {
	rt := router.GetRoute(p.Header.Cmd)
	if rt == nil {
		return NewFlyError(ErrNotFound, nil)
	}
	return rt.emitPacket(ctx, p)
}
