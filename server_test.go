package gircd

import (
	"bufio"
	"io"
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	hp := "localhost:12000"
	mb := NewMessageBus(10)
	go listenAndServe(hp, mb)
	time.Sleep(time.Second)
	client(t, hp)
}

func client(t *testing.T, hp string) {
	c, err := net.Dial("tcp", hp)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(c, "user a b c :d\r\n")
	io.WriteString(c, "nick test1\r\n")
	s := bufio.NewScanner(c)
	for s.Scan() {
		t.Log(s.Text())
	}
}
