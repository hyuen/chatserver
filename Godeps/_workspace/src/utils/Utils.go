package utils

import (
	"crypto/sha1"
	"crypto/rand"
	"fmt"
)

func SHA1(s string) string {
	output := fmt.Sprintf("%x", sha1.Sum([]byte(s)))
	return output
}

// generates a random string of the size of a SHA1 digest
func RandomSHA1() string {
	b := make([]byte, 60)
	rand.Read(b)
	return SHA1(fmt.Sprintf("%x", b))
}
	
