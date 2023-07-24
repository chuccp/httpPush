package user

type Store struct {
}

func newStore() *Store {
	return &Store{}
}

func (store *Store) AddUser(user IUser) bool {
	return false
}
func (store *Store) DeleteUser(user IUser) bool {
	return false
}
