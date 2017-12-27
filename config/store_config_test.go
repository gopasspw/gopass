package config

import "testing"

func TestStoreConfigMap(t *testing.T) {
	sc := &StoreConfig{}
	scm := sc.ConfigMap()
	t.Logf("map: %+v", scm)
	if scm["nopager"] != "false" {
		t.Errorf("nopager should be false")
	}
}
