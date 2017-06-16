package root

// GitInit initializes the git repo
func (r *Store) GitInit(name, sk string) error {
	store := r.getStore(name)
	return store.GitInit(store.Alias(), sk)
}

// Git runs arbitrary git commands on this store and all substores
func (r *Store) Git(name string, args ...string) error {
	store := r.getStore(name)
	return store.Git(args...)
}
