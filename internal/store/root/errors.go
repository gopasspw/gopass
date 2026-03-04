package root

import "fmt"

// AlreadyMountedError is an error that is returned when
// a store is already mounted on a given mount point.
type AlreadyMountedError string

func (a AlreadyMountedError) Error() string {
	// important: must pass a as string(a)!
	return fmt.Sprintf("%s is already mounted", string(a))
}

// NotInitializedError is an error that is returned when
// a not initialized store should be mounted.
type NotInitializedError struct {
	alias string
	path  string
}

// Alias returns the store alias this error was generated for.
func (n NotInitializedError) Alias() string { return n.alias }

// Path returns the store path this error was generated for.
func (n NotInitializedError) Path() string { return n.path }

func (n NotInitializedError) Error() string {
	return fmt.Sprintf("password store %s is not initialized. Try gopass init --store %s --path %s", n.alias, n.alias, n.path)
}
