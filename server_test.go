package gircd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

func Testserver(t *testing.T) {
	hp := "localhost:12000"
	mb := newMessageBus(10)
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

var n int
var l sync.RWMutex

func get() int {
	l.RLock()
	defer l.RUnlock()
	return n
}

func set(v int) {
	l.Lock()
	n = v
	l.Unlock()
}

const (
	max   = 20000
	stopn = 200000
)

func TestconditionVar(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var m sync.Mutex
	c := sync.NewCond(&m)
	resChan := make(chan int)

	ans := 0
	for i := 0; i < stopn; i++ {
		ans += i
	}
	go func() {
		tcnt := 0
		for v := range resChan {
			tcnt++
			if v != ans {
				panic("")
			}
		}
		fmt.Println(tcnt)
	}()
	var wg sync.WaitGroup
	wg.Add(max)
	for i := 0; i < max; i++ {
		go func() {
			cur := 0
			cnt := 0
			n2 := 0
			for {
				m.Lock()
				n1 := get()
				if n1 == n2 { // didn't change while we were working
					c.Wait()
				}
				n2 = get()
				n := n2
				m.Unlock()
				for i := cur; i < n; i++ {
					// process cur ... n
					cnt += i
				}
				cur = n
				if n == stopn {
					resChan <- cnt
					wg.Done()
					return
				}
			}
		}()
	}
	go func() {
		for i := 0; i < stopn; i++ {
			set(i)
			m.Lock()
			c.Broadcast()
			m.Unlock()
		}
		set(stopn)
		m.Lock()
		c.Broadcast()
		m.Unlock()
	}()
	/*
		go func() {
			for {
			}
		}()
	*/
	wg.Wait()
	close(resChan)
}

func TestcloseChan(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	closeChanFactory := make(chan chan struct{})
	closeChanSignal := make(chan struct{})

	go func(cc chan chan struct{}, nc chan struct{}) {
		c := make(chan struct{})
		for {
			select {
			case cc <- c:
			case <-nc:
				close(c)
				c = make(chan struct{})
			}
		}
	}(closeChanFactory, closeChanSignal)

	resChan := make(chan int)
	ans := 0
	for i := 0; i < stopn; i++ {
		ans += i
	}
	go func() {
		tcnt := 0
		for v := range resChan {
			tcnt++
			if v != ans {
				panic("")
			}
		}
		fmt.Println(tcnt)
	}()

	var wg sync.WaitGroup
	wg.Add(max)
	for i := 0; i < max; i++ {
		go func(cf chan chan struct{}) {
			cur := 0
			cnt := 0
			n2 := 0
			for {
				c := <-cf
				n1 := get()
				if n1 == n2 {
					<-c
				}
				n2 := get()
				n := n2
				for i := cur; i < n; i++ {
					cnt += i
				}
				cur = n
				if n >= stopn {
					resChan <- cnt
					wg.Done()
					return
				}
			}
		}(closeChanFactory)
	}

	go func() {
		for i := 0; i < stopn; i++ {
			set(i)
			i++
			// signal all observers using close(chan)
			closeChanSignal <- struct{}{}
		}
		set(stopn)
		closeChanSignal <- struct{}{}
	}()
	wg.Wait()
	close(resChan)
}

func TestconditionVarBroadcast(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var m sync.Mutex
	c := sync.NewCond(&m)
	go func() {
		for {
			m.Lock()
			c.Wait()
			m.Unlock()
			fmt.Println("a")
			time.Sleep(time.Second * 4)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second * 4)
			m.Lock()
			c.Wait()
			m.Unlock()
			fmt.Println("b")
		}
	}()

	go func() {
		for {
			m.Lock()
			c.Wait()
			m.Unlock()
			fmt.Println("c")
		}
	}()

	for i := 0; i < 10; i++ {
		m.Lock()
		c.Broadcast()
		fmt.Println("broadcast")
		m.Unlock()
		time.Sleep(time.Second * 2)
	}
	time.Sleep(time.Second * 30)
}
