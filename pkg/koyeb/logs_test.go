package koyeb

import "testing"

// Stop must be safe to call more than once and after the ping goroutine
// already exited on its own.
func TestWebsocketPingConnectionStopIsIdempotent(t *testing.T) {
	conn := NewWebsocketPingConnection(nil)
	conn.Stop()
	conn.Stop()
}
