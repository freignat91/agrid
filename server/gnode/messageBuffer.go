package gnode

import (
	"sync"
)

type MessageBuffer struct {
	maxSize       int
	overflowLimit int
	size          int
	values        []*AntMes
	in            int
	out           int
	lock          sync.RWMutex
	ioChan        chan string
	max           int //For stats
	overflow      bool
}

func (m *MessageBuffer) init(size int) {
	m.maxSize = size
	m.overflowLimit = (size * 8) / 10
	m.values = make([]*AntMes, size, size)
	m.ioChan = make(chan string)
	m.lock = sync.RWMutex{}
	m.overflow = false
}

func (m *MessageBuffer) get(wait bool) (*AntMes, bool) {
	//logf.info("BufferGet in=%d out=%d size=%d\n", m.in, m.out, m.size)
	if m.size == 0 {
		if wait {
			m.ioChan <- "ok"
		} else {
			return nil, false
		}
	}
	m.lock.Lock()
	mes := m.values[m.out]
	m.values[m.out] = nil
	m.out = m.incrIndex(m.out)
	m.size--
	if m.size < m.overflowLimit {
		m.overflow = false
		if m.size == 0 {
			m.in = 0
			m.out = 0
		}
	}
	m.lock.Unlock()
	return mes, true
}

func (m *MessageBuffer) put(mes *AntMes) bool {
	if m.overflow {
		return false
	}
	m.lock.Lock()
	//logf.info("BufferPut in=%d out=%d size=%d\n", m.in, m.out, m.size)
	if m.size >= m.maxSize {
		m.overflow = true
		m.lock.Unlock()
		return false
	}
	m.values[m.in] = mes
	m.in = m.incrIndex(m.in)
	m.size++
	if m.size > m.max {
		m.max = m.size
	}
	select {
	case <-m.ioChan:
	default:
	}
	m.lock.Unlock()
	return true
}

func (m *MessageBuffer) incrIndex(index int) int {
	index++
	if index >= m.maxSize {
		index = 0
	}
	return index
}

func (m *MessageBuffer) Clear() {
	m.lock.Lock()
	for i, _ := range m.values {
		m.values[i] = nil
	}
	m.in = 0
	m.out = 0
	m.size = 0
	m.lock.Unlock()
}

func (m *MessageBuffer) isAvailable() bool {
	if m.size >= m.maxSize {
		return false
	}
	return true
}
