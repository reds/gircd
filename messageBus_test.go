package gircd

import (
	"runtime"
	"strconv"
	"sync"
	"testing"
)

var res = make([][]int, 10000)
var resLock sync.Mutex
var wg sync.WaitGroup

func a(mb *messageBus, n int) {
	curn := 0
	for {
		mb.wait()
		for {
			v := mb.get(curn)
			if v == nil {
				// fmt.Println("v nil")
				break
			}
			if curn == len(res) {
				wg.Done()
				return
			}
			nn, _ := strconv.Atoi(string(v))
			res[n] = append(res[n], nn)
			curn++
		}
	}
}

func TestMB(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	wg.Add(len(res))
	mb := newMessageBus(2000)
	for i := 0; i < len(res); i++ {
		go a(mb, i)
	}
	go func() {
		n := 0
		for {
			mb.appendRaw([]byte(strconv.Itoa(n)))
			n++
		}
	}()
	wg.Wait()
	for i := 0; i < len(res); i++ {
		for i, v := range res[i] {
			if i != v {
				t.Fatal(i, v, res)
			}
		}
	}
}
