package ws

import (
	"context"
	"errors"
	"github.com/MikhailGulkin/packages/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Manager struct {
	upgrader        websocket.Upgrader
	processorFabric PipeProcessorFabric
	clients         map[string]Client

	isClosed   atomic.Bool
	deadSignal chan string
	mu         sync.Mutex
	logger     Logger
}

func NewManager(
	opts ...OptionFunc,
) *Manager {
	manager := &Manager{
		upgrader:        websocket.Upgrader{},
		processorFabric: &PipeProcessorFabricImpl{},
		clients:         make(map[string]Client, defaultConnsLimit),
		deadSignal:      make(chan string, defaultConnsLimit),
		logger:          log.Default(),
		mu:              sync.Mutex{},
		isClosed:        atomic.Bool{},
	}
	return manager.With(opts...)
}

func (m *Manager) Process(uniqueID string, w http.ResponseWriter, r *http.Request, header http.Header) error {
	if m.isClosed.Load() {
		return ErrManagerClosed
	}

	conn, err := m.upgrader.Upgrade(w, r, header)
	if err != nil {
		return err
	}

	c := make(chan error)
	go func() {
		processor, err := m.processorFabric.NewPipeProcessor(r.Context(), uniqueID)
		if err != nil {
			m.logger.Errorw("error creating pipe processor", "error", err)
			c <- err
			return
		}
		defer func() {
			err := processor.Close()
			if err != nil {
				m.logger.Errorw("error closing pipe processor", "error", err)
			}
		}()

		client := NewDefaultClient(conn, uuid.New().String(), m.deadSignal, processor, m.logger)
		m.addClient(client.GetClientID(), client)

		err = client.Run(context.WithoutCancel(r.Context()))
		if err != nil {
			m.logger.Errorw("Client run error", "clientID", client.GetClientID(), "error", err)
		}
	}()
	select {
	case <-r.Context().Done():
		return nil
	case err := <-c:
		return errors.Join(err, conn.Close())
	case <-time.After(connCreateTimeout):
		return ErrCreateConnTimeout
	}
}
func (m *Manager) addClient(id string, client Client) {
	defer m.mu.Unlock()
	m.mu.Lock()
	m.clients[id] = client
}
func (m *Manager) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case id, ok := <-m.deadSignal:
			if !ok {
				return
			}
			func() {
				defer m.mu.Unlock()
				m.mu.Lock()
				_, ok := m.clients[id]
				if !ok {
					m.logger.Errorw("DefaultClient not found", "id", id)
					return
				}
				delete(m.clients, id)
			}()
		}
	}
}

func (m *Manager) Close() error {
	defer m.mu.Unlock()
	m.mu.Lock()

	var err error
	m.isClosed.Store(true)
	for id, client := range m.clients {
		err := client.Close()
		if err != nil {
			m.logger.Errorw("error closing connection", "error", err)
			err = errors.Join(err, err)
		}
		delete(m.clients, id)
	}
	close(m.deadSignal)
	return err
}
