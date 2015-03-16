package gircd

import "sync"

func (mb *messageBus) appendRaw(buf []byte) error {
	nbuf := make([]byte, len(buf))
	copy(nbuf, buf)
	mb.Lock()
	mb.qend++
	mb.dataBlocks[mb.qend%mb.qlen] = nbuf
	mb.closeChanSignal <- struct{}{}
	mb.Unlock()
	return nil
}

func (mb *messageBus) get(start int) [][]byte {
	mb.RLock()
	res := make([][]byte, 0, mb.qend-start)
	for i := start; i <= mb.qend; i++ {
		buf := mb.dataBlocks[i%mb.qlen]
		nbuf := make([]byte, len(buf))
		copy(nbuf, buf)
		res = append(res, nbuf)
	}
	mb.RUnlock()
	return res
}

type messageType int

const (
	messageTypeRaw = iota
)

type messageBus struct {
	sync.RWMutex
	// cirular list of data blocks
	dataBlocks       [][]byte
	qlen             int
	qend             int
	typ              messageType
	closeChanFactory chan chan struct{}
	closeChanSignal  chan struct{}
}

func NewMessageBus(qlen int) *messageBus {
	mb := &messageBus{dataBlocks: make([][]byte, qlen), qlen: qlen,
		closeChanFactory: make(chan chan struct{}),
		closeChanSignal:  make(chan struct{}),
	}
	go func(cc chan chan struct{}, nc chan struct{}) {
		c := make(chan struct{})
		for {
			select {
			case cc <- c:
				//fmt.Println("Sending wait chan")
			case <-nc:
				close(c)
				c = make(chan struct{})
			}
		}
	}(mb.closeChanFactory, mb.closeChanSignal)
	return mb
}
