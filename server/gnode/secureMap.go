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
	if value == nil {
		return
	}
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
	val, ok := m.objectMap[key]
	if val == nil {
		return false
	}
	return ok
}

func (m *secureMap) del(key string) {
	m.lock.Lock()
	delete(m.objectMap, key)
	m.lock.Unlock()
}

func (m *secureMap) len() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.objectMap)
}

func (m *secureMap) clear() {
	m.objectMap = make(map[string]interface{})
}
