package gnode

import (
	"sync"
	"time"
)

type gnodeIdMap struct {
	valueMap map[string]time.Time
	lock     sync.RWMutex
}

func (m *gnodeIdMap) Init() {
	m.lock = sync.RWMutex{}
	m.valueMap = make(map[string]time.Time)
	m.startBackgroundCleanUp()
}

func (m *gnodeIdMap) Add(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.valueMap[id] = time.Now()
}

func (m *gnodeIdMap) Len() int {
	return len(m.valueMap)
}

func (m *gnodeIdMap) Exists(id string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.valueMap[id]
	if ok {
		m.valueMap[id] = time.Now()
	}
	return ok
}

func (m *gnodeIdMap) CleanUp() {
	m.lock.RLock()
	defer m.lock.RUnlock()
	//logf.info("Start idMap cleanup: %d\n", m.Len())
	now := time.Now()
	for id, lastTime := range m.valueMap {
		if now.Sub(lastTime).Seconds() > 10 {
			delete(m.valueMap, id)
		}
	}
	//logf.info("idMap cleanup: %d\n", m.Len())

}

func (m *gnodeIdMap) startBackgroundCleanUp() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			m.CleanUp()
		}
	}()
}
