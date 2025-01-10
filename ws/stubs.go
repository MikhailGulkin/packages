package ws

import (
	"context"
	"fmt"
)

type ProcessorImpl struct {
	send chan []byte
}

func NewProcessorImpl() *ProcessorImpl {
	return &ProcessorImpl{
		send: make(chan []byte, 256),
	}
}

func (p *ProcessorImpl) ProcessRead(ctx context.Context, messageType int, msg []byte) ([]byte, error) {
	fmt.Println("ProcessRead", messageType, msg)
	return msg, nil
}

func (p *ProcessorImpl) ListenWrite() <-chan []byte {
	return p.send
}

type PipeProcessorFabricImpl struct{}

func (p *PipeProcessorFabricImpl) NewPipeProcessor(ctx context.Context, userID string) (PipeProcessor, error) {
	return NewProcessorImpl(), nil
}
