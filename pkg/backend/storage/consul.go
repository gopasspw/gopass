// +build consul

package storage

import _ "github.com/gopasspw/gopass/pkg/backend/storage/kv/consul" // registers consul backend if build tag consul is set
