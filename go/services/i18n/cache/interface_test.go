package cache

import "testing"

func TestCache_Interface(t *testing.T) {
	var _ Cache = (*MockCache)(nil)
}
