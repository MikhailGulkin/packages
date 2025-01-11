package ws

import (
	"context"
	"fmt"
	"time"
)

type ProcessorImpl struct {
	r ReadPipeProcessor
	w WritePipeProcessor
}

func (p *ProcessorImpl) Close() error {
	return nil
}

func NewProcessorImpl(
	ReadProcessor ReadPipeProcessor,
	WriteProcessor WritePipeProcessor,
) *ProcessorImpl {
	return &ProcessorImpl{
		r: ReadProcessor,
		w: WriteProcessor,
	}
}

type ReadPipeProcessorImpl struct{}

func (r *ReadPipeProcessorImpl) ProcessRead(ctx context.Context, messageType int, msg []byte) ([]byte, error) {
	fmt.Println("ProcessRead", messageType, msg)
	return msg, nil
}

type WritePipeProcessorImpl struct {
	send chan []byte
}

func NewWritePipeProcessorImpl() *WritePipeProcessorImpl {
	send := make(chan []byte, 256)
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				send <- []byte("some msg default listen write")
			}
		}
	}()

	return &WritePipeProcessorImpl{
		send: send,
	}
}

func (w *WritePipeProcessorImpl) ListenWrite(ctx context.Context) <-chan []byte {
	return w.send
}

func (p *ProcessorImpl) ProcessRead(ctx context.Context, messageType int, msg []byte) ([]byte, error) {
	return p.r.ProcessRead(ctx, messageType, msg)
}

func (p *ProcessorImpl) ListenWrite(ctx context.Context) <-chan []byte {
	return p.w.ListenWrite(ctx)
}

type PipeProcessorFabricImpl struct{}

func (p *PipeProcessorFabricImpl) Close() error {
	//TODO implement me
	panic("implement me")
}

func (p *PipeProcessorFabricImpl) NewPipeProcessor(ctx context.Context, userID string) (PipeProcessor, error) {
	return NewProcessorImpl(
		&ReadPipeProcessorImpl{},
		NewWritePipeProcessorImpl(),
	), nil
}
