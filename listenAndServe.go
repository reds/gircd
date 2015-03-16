package gircd

import (
	"bufio"
	"fmt"
	"net"
)

func serve(c net.Conn, mb *messageBus) {
	go func() {
		s := bufio.NewScanner(c)
		for s.Scan() {
			err := mb.appendRaw(s.Bytes())
			if err != nil {
				return
			}
		}
	}()
	curmsg := 0
	for {
		sig := <-mb.closeChanFactory
		<-sig
		fmt.Println("got message")
		for _, v := range mb.get(curmsg) {
			fmt.Println("msg", string(v))
			curmsg++
			_, err := c.Write(v)
			if err != nil {
				return
			}
		}
	}
}

func listenAndServe(hp string, mb *messageBus) {
	l, err := net.Listen("tcp", hp)
	if err != nil {
		panic(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go serve(c, mb)
	}
}
