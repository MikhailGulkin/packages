package ws

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
	"sync"
	"sync/atomic"
	"time"
)

type DefaultClient struct {
	conn *websocket.Conn
	id   string

	deadSignal chan string
	closeChan  chan error

	pipeProcessor PipeProcessor
	logger        Logger

	close atomic.Bool
	mu    sync.Mutex
}

func NewDefaultClient(
	conn *websocket.Conn,
	id string,
	deadSignal chan string,
	pipeProcessor PipeProcessor,
	logger Logger,
) *DefaultClient {
	return &DefaultClient{
		conn:          conn,
		id:            id,
		pipeProcessor: pipeProcessor,
		deadSignal:    deadSignal,
		logger:        logger,
		closeChan:     make(chan error, 1),
		close:         atomic.Bool{},
		mu:            sync.Mutex{},
	}
}
func (c *DefaultClient) Configure() error {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetCloseHandler(func(code int, text string) error {
		c.logger.Infow("connection closed", "code", code, "text", text)
		return c.Close()

	})
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return err
	}
	c.conn.SetPongHandler(func(appData string) error {
		c.logger.Infow("pong received", "appData", appData)
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			c.logger.Errorw("error setting read deadline", "error", err)
			return err
		}
		return nil
	})

	return nil

}

func (c *DefaultClient) Run(ctx context.Context) error {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			c.logger.Errorw("error closing connection", "error", err)
		}
		c.deadSignal <- c.GetClientID()
	}()

	if err := c.Configure(); err != nil {
		return err
	}

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case err := <-c.closeChan:
			return err
		}
	})
	errGroup.Go(func() error {
		return c.ReadPipe(ctx)
	})
	errGroup.Go(func() error {
		return c.WritePipe(ctx)
	})
	errGroup.Go(func() error {
		return c.Ping(ctx)
	})
	return errGroup.Wait()
}

func (c *DefaultClient) ReadPipe(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.Infow("ReadPipe ctx done", "clientID", c.GetClientID())
			return nil
		default:
			messageType, msg, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return ErrCloseProperly
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					return errors.Join(err, ErrConnectionCloseIncorrect)
				}
				return errors.Join(err, ErrUnknownReadException)
			}
			answer, err := c.pipeProcessor.ProcessRead(ctx, messageType, msg)
			if err != nil {
				return errors.Join(err, ErrProcessRead)
			}

			if len(answer) == 0 {
				return nil
			}

			err = c.conn.WriteMessage(websocket.TextMessage, answer)
			if err != nil {
				return errors.Join(err, ErrWriteAnswer)
			}
		}
	}
}

func (c *DefaultClient) WritePipe(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.Infow("WritePipe ctx done", "ctxErr", ctx.Err(), "id", c.GetClientID())
			return nil
		case msg, ok := <-c.pipeProcessor.ListenWrite(ctx):
			if !ok {
				return nil
			}
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return err
			}

			err = c.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				return errors.Join(err, ErrWriteAnswer)
			}
		}
	}
}

func (c *DefaultClient) Ping(ctx context.Context) error {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.logger.Infow("Ping ctx done", "ctxErr", ctx.Err(), "clientID", c.GetClientID())
			return nil
		case <-ticker.C:
			err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
			if err != nil {
				return errors.Join(err, ErrWriteAnswer)
			}
		}
	}
}

func (c *DefaultClient) Close() error {
	defer c.mu.Unlock()
	c.mu.Lock()
	if c.close.Load() {
		return nil
	}

	c.close.Store(true)
	c.closeChan <- ErrCloseProperly
	close(c.closeChan)

	return c.conn.Close()
}

func (c *DefaultClient) GetClientID() string {
	return c.id
}
