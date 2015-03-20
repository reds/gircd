package gircd

import "sync"

func (mb *messageBus) appendRaw(buf []byte) error {
	nbuf := make([]byte, len(buf))
	copy(nbuf, buf)
	mb.Lock()
	mb.dataBlocks[mb.qend%mb.qlen] = nbuf
	mb.qend++
	mb.cMutex.Lock()
	mb.cVar.Broadcast()
	mb.cMutex.Unlock()
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
	dataBlocks [][]byte
	qlen       int
	qend       int
	cMutex     *sync.Mutex
	cVar       *sync.Cond
}

func newMessageBus(qlen int) *messageBus {
	m := &sync.Mutex{}
	mb := &messageBus{
		dataBlocks: make([][]byte, qlen),
		qlen:       qlen,
		cMutex:     m,
		cVar:       sync.NewCond(m),
	}
	return mb
}

func (mb *messageBus) wait() {
	mb.cMutex.Lock()
	mb.cVar.Wait()
	mb.cMutex.Unlock()
}
