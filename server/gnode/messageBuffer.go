package gnode

import (
	"sync"
)

type MessageBuffer struct {
	maxSize int
	size    int
	values  []*AntMes
	in      int
	out     int
	lock    sync.RWMutex
	ioChan  chan string
	max     int //For stats
}

func (m *MessageBuffer) init(size int) {
	m.maxSize = size
	m.values = make([]*AntMes, size, size)
	m.ioChan = make(chan string)
	m.lock = sync.RWMutex{}
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
	defer m.lock.Unlock()
	mes := m.values[m.out]
	m.values[m.out] = nil
	m.out = m.incrIndex(m.out)
	m.size--
	if m.size == 0 {
		m.in = 0
		m.out = 0
	}
	return mes, true
}

func (m *MessageBuffer) put(mes *AntMes) bool {
	m.lock.Lock()
	//logf.info("BufferPut in=%d out=%d size=%d\n", m.in, m.out, m.size)
	defer m.lock.Unlock()
	if m.size >= m.maxSize {
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
	defer m.lock.Unlock()
	for i, _ := range m.values {
		m.values[i] = nil
	}
	m.in = 0
	m.out = 0
	m.size = 0
}
