package neptulon

import (
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"time"
)

// GenID generates a unique ID using crypto/rand in the form of "m-96bitBase16" and total of 26 characters long (i.e. m-18dc2ae3898820d9c5df4f38).
func GenID() (string, error) {
	// todo: we can use sequential numbers optionally, just as the Android client does (1, 2, 3..) in upstream messages
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("m-%x", b), nil
}

var letters = []rune(". !abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// randString generates a random string sequence of given size.
func randString(n int) string {
	mathrand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[mathrand.Intn(len(letters))]
	}
	return string(b)
}
