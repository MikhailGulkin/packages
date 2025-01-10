package ws

import "errors"

var (
	ErrManagerClosed            = errors.New("manager is closed")
	ErrCloseProperly            = errors.New("error close properly")
	ErrProcessRead              = errors.New("error process read")
	ErrConnectionCloseIncorrect = errors.New("connection closed incorrect")
	ErrUnknownReadException     = errors.New("unknown read exception")
	ErrWriteAnswer              = errors.New("error write answer")
)
