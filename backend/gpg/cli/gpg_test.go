package cli

import "testing"

func TestSplitPacket(t *testing.T) {
	m := splitPacket(":pubkey enc packet: version 3, algo 16, keyid 6780DF473C7A71D3")
	val, found := m["keyid"]
	if !found {
		t.Errorf("Failed to parse/lookup keyid")
	}
	if val != "6780DF473C7A71D3" {
		t.Errorf("Failed to get keyid")
	}
}
