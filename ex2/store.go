package ex2

import (
	"sync"
)

type Store struct {
	clientMap *sync.Map
}

func NewStore() *Store {
	return &Store{clientMap: new(sync.Map)}
}
func (s *Store) LoadOrStore(cl *client, username string) (*client, bool) {
	v, ok := s.clientMap.LoadOrStore(username, cl)
	return v.(*client), ok
}

func (s *Store) Delete(username string) {
	s.clientMap.Delete(username)
}

func (s *Store) RangeClient(f func(c *client)) {
	s.clientMap.Range(func(key, value any) bool {
		f(value.(*client))
		return true
	})
}
