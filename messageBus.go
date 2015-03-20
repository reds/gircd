package gircd

import "sync"

func (mb *messageBus) appendRaw(buf []byte) error {
	nbuf := make([]byte, len(buf))
	copy(nbuf, buf)
	mb.Lock()
	mb.dataBlocks[mb.qend%mb.qlen] = nbuf
	mb.qend++
	mb.closeChanSignal <- struct{}{}
	mb.Unlock()
	return nil
}

func (mb *messageBus) get(msgnum int) []byte {
	mb.RLock()
	defer mb.RUnlock()
	if msgnum < mb.qend {
		return mb.dataBlocks[msgnum%mb.qlen]
	}
	return nil
}

type messageBus struct {
	sync.RWMutex
	// cirular list of data blocks
	dataBlocks       [][]byte
	qlen             int
	qend             int
	closeChanFactory chan chan struct{}
	closeChanSignal  chan struct{}
}

func newMessageBus(qlen int) *messageBus {
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

func (mb *messageBus) wait() {
	sig := <-mb.closeChanFactory
	<-sig
}
