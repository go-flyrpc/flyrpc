package flyrpc

import (
	"log"
	"reflect"
	"runtime/debug"
	"strings"
)

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
	inTypes     []reflect.Type
	outTypes    []reflect.Type
	outErrIndex int
	outType     reflect.Type
}

var (
	_err        error
	typeError   = reflect.TypeOf(&_err).Elem()
	typeContext = reflect.TypeOf(&Context{})
	typeBytes   = reflect.TypeOf([]byte{})
	typePacket  = reflect.TypeOf(&Packet{})
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
	r.inTypes = make([]reflect.Type, numIn)
	r.outTypes = make([]reflect.Type, numOut)
	for i := 0; i < numIn; i++ {
		r.inTypes[i] = r.vHandler.Type().In(i)
	}
	for i := 0; i < numOut; i++ {
		r.outTypes[i] = r.vHandler.Type().Out(i)
	}
	if numOut > 2 {
		panic("Too much returns, handler must return (Message, error) or error")
	}
	if numOut == 2 {
		if !r.outTypes[1].AssignableTo(typeError) {
			panic("Handler should return (Message, error)")
		}
	}
	if numOut > 0 {
		if r.outTypes[numOut-1].AssignableTo(typeError) {
			r.outErrIndex = numOut - 1
		}
		if !r.outTypes[0].AssignableTo(typeError) {
			r.outType = r.outTypes[0]
		}
	}
	return r
}

func (route *route) call(values []reflect.Value) (result []reflect.Value, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = NewFlyError(ErrHandlerPanic)
			lines := strings.Split(string(debug.Stack()), "\n")
			stack := strings.Join(lines[6:], "\n")
			log.Println("Handler panic: %s %s", r, stack)
		}
	}()
	result = route.vHandler.Call(values)
	return
}

func (route *route) emitPacket(ctx *Context, pkt *Packet) error {
	values := make([]reflect.Value, route.numIn)
	for i := 0; i < route.numIn; i++ {
		inType := route.inTypes[i]
		if inType == typeContext {
			values[i] = reflect.ValueOf(ctx)
		} else if inType == typeBytes {
			values[i] = reflect.ValueOf(pkt.MsgBuff)
		} else if inType == typePacket {
			values[i] = reflect.ValueOf(pkt)
		} else {
			v := reflect.New(inType.Elem())
			err := route.serializer.Unmarshal(pkt.MsgBuff, v.Interface())
			if err != nil {
				return err
			}
			values[i] = v
		}
	}
	ret, err := route.call(values)
	if err != nil {
		return err
	}
	// retSize := len(ret)
	// if retSize != route.numOut {
	// 	panic("Error result size")
	// }
	if route.outErrIndex >= 0 {
		ve := ret[route.outErrIndex]
		if !ve.IsNil() {
			err := ve.Interface().(error)
			if err != nil {
				flyErr, ok := err.(*flyError)
				if ok && flyErr.Code < 20000 {
					// client error
					return ctx.SendError(flyErr)
				}
				return err
			}
		}
	}
	if pkt.Header.Flag&RPCFlagReq == 0 {
		// not a call, no response
		return nil
	}
	if route.outType != nil {
		// rpc return
		vout := ret[0]
		bytes, err := route.serializer.Marshal(vout.Interface())
		if err != nil {
			return err
		}
		return ctx.SendPacket(
			TypeRPC|RPCFlagResp,
			pkt.Header.Cmd,
			pkt.Header.Seq,
			bytes)
	}
	// just return an empty ack message
	return ctx.SendPacket(
		TypeRPC|RPCFlagResp,
		pkt.Header.Cmd,
		pkt.Header.Seq,
		[]byte{},
	)
	return nil
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
