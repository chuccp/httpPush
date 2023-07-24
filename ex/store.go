package ex

import "sync"

type Store struct {
	clientMap *sync.Map
}

func NewStore() *Store {
	return &Store{clientMap: new(sync.Map)}
}
func (s *Store) LoadOrStore(c *client) (*client, bool) {
	v, ok := s.clientMap.LoadOrStore(c.username, c)
	return v.(*client), ok
}
