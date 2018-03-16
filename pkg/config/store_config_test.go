package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreConfigMap(t *testing.T) {
	sc := &StoreConfig{}
	assert.Equal(t, "false", sc.ConfigMap()["nopager"])
}
