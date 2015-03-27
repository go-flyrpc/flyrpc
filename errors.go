package fly

import "strconv"

const (
	// 10000 - 20000 client error
	ERR_NOT_FOUND    int = 10000
	ERR_BUFF_TO_LONG int = 11000
	// 20000 + server error
	ERR_NO_WRITER     int = 21000
	ERR_WRITER_CLOSED int = 21001
	// 25000 + serializer error
	ERR_NOT_PROTO_MESSAGE int = 25010
)

var messages map[int]string = map[int]string{
	ERR_NO_WRITER:     "NO_WRITER",
	ERR_WRITER_CLOSED: "WRITER_CLOSED",
}

type FlyError struct {
	Code    int
	Message string
	Err     error
}

func (f *FlyError) Error() string {
	return "FlyError " + strconv.Itoa(f.Code) + " " + f.Message
}

func NewFlyError(code int, args ...error) *FlyError {
	var err error
	if len(args) > 0 {
		err = args[0]
	}
	return &FlyError{
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
