/*
FlyRPC provide a flexiable way to communicate between Server and Client.

It support JSON, Msgpack, Protobuf serializer.

It support Call/Response  or Send/Receive pattern.
*/
package flyrpc

type Error interface {
	Code() string
	Error() string
}

const (
	// Common error
	ErrTimeOut string = "TIMEOUT"

	// 10000 - 20000 client error

	ErrNotFound       string = "NOT_FOUND"
	ErrUnknownSubType string = "UNKNOWN_SUB_TYPE"
	ErrBuffTooLong    string = "BUFF_TOO_LONG"
	// 20000 + server error

	ErrNoWriter     string = "NO_WRITER"
	ErrWriterClosed string = "WRITER_CLOSED"
	ErrHandlerPanic string = "HANDLER_PANIC"
	// 25000 + serializer error

	ErrNotProtoMessage string = "NOT_PROTOBUF_MESSAGE"
)

type flyError struct {
	code string
	Err  error
}

func (f *flyError) Error() string {
	return f.code
}

func (f *flyError) Code() string {
	return f.code
}

func NewFlyError(code string, args ...error) *flyError {
	var err error
	if len(args) > 0 {
		err = args[0]
	}
	return &flyError{
		code: code,
		Err:  err,
	}
}

type ReplyError struct {
	flyError
	pkt *Packet
}

func newReplyError(code string, pkt *Packet) *ReplyError {
	return &ReplyError{
		flyError: flyError{code: code},
		pkt:      pkt,
	}
}
