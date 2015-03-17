package gircd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type nicks struct {
	sync.RWMutex
	nicks map[string]*user
}

var allNicks = &nicks{nicks: make(map[string]*user)}

type user struct {
	username  string
	nick      string
	password  string
	registerd bool
	realname  string
}

func (ns *nicks) get(u *user) *user {
	ns.RLock()
	defer ns.RUnlock()
	if eu, exists := ns.nicks[u.nick]; exists {
		return eu
	}
	return nil
}

func serve(c net.Conn, mb *messageBus) {
	defer c.Close()
	s := bufio.NewScanner(c)
	// authenticate: user, nick, pass
	u, err := auth(s)
	if err != nil {
		log.Println(err)
		return
	}
	if n := allNicks.get(u); n != nil {
		log.Println("nick already exists", n)
		io.WriteString(c, "Nick already exists\r\n")
		return
	}
	go func() {
		for s.Scan() {
			cmd, err := ircParse(s.Text())
			if err != nil {
				return
			}
			fmt.Println(cmd)
		}
	}()
	curmsg := 0
	for {
		sig := <-mb.closeChanFactory
		<-sig
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
