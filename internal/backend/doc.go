// Package backend implements a registry to register differnet plugable backends for encryption and storage (incl. version control).
// The actual backends are implemented in the subpackages. They register themselves in the registry with blank imports.
package backend
