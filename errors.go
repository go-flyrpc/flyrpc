/*
FlyRPC provide a flexiable way to communicate between Server and Client.

It support JSON, Msgpack, Protobuf serializer.

It support Call/Response  or Send/Receive pattern.
*/
package flyrpc

import "errors"

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

type ReplyError struct {
	code  string
	cause error
	pkt   *Packet
}

func (e *ReplyError) Error() string {
	return e.code
}

func newReplyError(code string, pkt *Packet) *ReplyError {
	return &ReplyError{
		code: code,
		pkt:  pkt,
	}
}

func newError(code string) error {
	return errors.New(code)
}

func newFlyError(code string, cause error) error {
	return &ReplyError{
		code:  code,
		cause: cause,
	}
}
