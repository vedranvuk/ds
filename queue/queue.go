package queue

import "sync"

type Queue[K any] struct {
	m     sync.Mutex
	items []K
}

func (self *Queue[K]) Push(v K) {
	self.m.Lock()
	self.items = append(self.items, v)
	self.m.Unlock()
}

func (self *Queue[K]) Pop() (v K, b bool) {
	self.m.Lock()
	if len(self.items) == 0 {
		self.m.Unlock()
		return v, false
	}
	v = self.items[0]
	self.items = self.items[1:]
	self.m.Unlock()
	return v, true
}
