package safemap

import (
	"sync"
)

type SafeMap struct {
	valueMap map[string]string
	lock     sync.RWMutex
}

func (m *SafeMap) Init() {
	m.lock = sync.RWMutex{}
	m.valueMap = make(map[string]string)
}

func (m *SafeMap) Get(key string) (string, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok := m.valueMap[key]
	return val, ok
}

func (m *SafeMap) Set(key string, val string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.valueMap[key] = val
}

func (m *SafeMap) Len() int {
	return len(m.valueMap)
}

func (m *SafeMap) GetValueMap() map[string]string {
	return m.valueMap
}

func (m *SafeMap) Exists(key string) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *SafeMap) Clear() {
	m.valueMap = make(map[string]string)
}
