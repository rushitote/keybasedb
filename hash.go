package main

import (
	"crypto/sha256"
)

// Generates a SHA256 hash of a string
func GenerateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return string(h.Sum(nil))
}
