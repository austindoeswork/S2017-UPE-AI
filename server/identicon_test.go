package server

import (
	"fmt"
	"testing"
)

func TestNewIdenticon(t *testing.T) {
	hash := "thaitai"
	fmt.Printf("Using hash: '%s' -> ", hash)
	hash = GenerateHash(hash)
	fmt.Printf("'%s'\n", hash)
	icon := NewIdenticon(hash, nil)
	icon.Save("test.png")
}
