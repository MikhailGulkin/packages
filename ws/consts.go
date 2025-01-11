package ws

import "time"

const (
	connCreateTimeout = 5 * time.Second
	// Time allowed to write a message to the peer.
	writeWait = 15 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize    = 512
	defaultConnsLimit = 100
)
