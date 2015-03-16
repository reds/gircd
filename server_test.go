package gircd

import "testing"

func TestServer(t *testing.T) {
	mb := NewMessageBus(10)
	listenAndServe("localhost:12000", mb)
}
