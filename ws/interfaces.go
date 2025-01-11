package ws

import "context"

type Logger interface {
	Infow(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

type PipeProcessor interface {
	ReadPipeProcessor
	WritePipeProcessor
}

type ReadPipeProcessor interface {
	ProcessRead(ctx context.Context, messageType int, msg []byte) ([]byte, error)
}

type WritePipeProcessor interface {
	ListenWrite(ctx context.Context) <-chan []byte
}

type PipeProcessorFabric interface {
	NewPipeProcessor(ctx context.Context, uniqueID string) (PipeProcessor, error)
}

type Client interface {
	GetClientID() string
	Run(ctx context.Context) error
	Close() error
}
