package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"os"
)

func grantPermission(username, password string) bool {
	correctUser := os.Getenv("BASIC_AUTH_USER")
	correctPass := os.Getenv("BASIC_AUTH_PASS")

	if correctUser == "" || correctPass == "" {
		log.Println("BASIC_AUTH_USER and BASIC_AUTH_PASS must be set")
		return false
	}

	return match(username, correctUser) && match(password, correctPass)
}

// match provides a safe way of comapring strings by using
// `subtle.ConstantTimeCompare()` to avoid timing attacks.
// Inputs are hashed to make sure we compare same-length arguments,
// to avoid leaking any information as comparing slices of different
// length is not done in constant time (it returns early).
func match(a, b string) bool {
	hashA := sha256.Sum256([]byte(a))
	hashB := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(hashA[:], hashB[:]) == 1
}
