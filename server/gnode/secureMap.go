package gnode

import (
	"sync"
)

type secureMap struct {
	objectMap map[string]interface{}
	lock      sync.RWMutex
}

func (m *secureMap) init() {
	m.objectMap = make(map[string]interface{})
	m.lock = sync.RWMutex{}
}

func (m *secureMap) set(key string, value interface{}) {
	m.lock.Lock()
	m.objectMap[key] = value
	m.lock.Unlock()
}

func (m *secureMap) get(key string) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.objectMap[key]
}

func (m *secureMap) exists(key string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.objectMap[key]
	return ok
}

func (m *secureMap) del(key string) {
	m.lock.Lock()
	delete(m.objectMap, key)
	m.lock.Unlock()
}

func (m *secureMap) len() int {
	return len(m.objectMap)
}
