/*
FlyRPC provide a flexiable way to communicate between Server and Client.

It support JSON, Msgpack, Protobuf serializer.

It support Call/Response  or Send/Receive pattern.
*/
package flyrpc

import "strconv"

const (
	// Common error
	ErrTimeOut int = 1000

	// 10000 - 20000 client error

	ErrNotFound       int = 10000
	ErrUnknownSubType int = 10001
	ErrBuffTooLong    int = 11000
	// 20000 + server error

	ErrNoWriter     int = 21000
	ErrWriterClosed int = 21001
	ErrHandlerPanic int = 22000
	// 25000 + serializer error

	ErrNotProtoMessage int = 25010
)

var messages = map[int]string{
	ErrTimeOut:         "TIMEOUT",
	ErrNotFound:        "NOT_FOUND",
	ErrBuffTooLong:     "BUFF_TOO_LONG",
	ErrNoWriter:        "NO_WRITER",
	ErrWriterClosed:    "WRITER_CLOSED",
	ErrNotProtoMessage: "NOT_PROTO_MESSAGE",
	ErrUnknownSubType:  "UNKNOWN_SUB_TYPE",
}

type flyError struct {
	Code    int
	Message string
	Err     error
}

func (f *flyError) Error() string {
	return "FlyError " + strconv.Itoa(f.Code) + " " + f.Message
}

func NewFlyError(code int, args ...error) *flyError {
	var err error
	if len(args) > 0 {
		err = args[0]
	}
	return &flyError{
		Code:    code,
		Message: messages[code],
		Err:     err,
	}
}

/*
type ClientError struct {
	FlyError
}

type ServerError struct {
	FlyError
}

func NewServerError(code int, msg string, err error) *ServerError {
	return &ServerError{*NewFlyError(code, msg, err)}
}

func NewClientError(code int, msg string, err error) *ClientError {
	return &ClientError{*NewFlyError(code, msg, err)}
}
*/
