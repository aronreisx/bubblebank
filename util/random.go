package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// RandomInt generates a random integer between minVal and max
func RandomInt(minVal, max int64) int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(max-minVal+1))
	return minVal + n.Int64()
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c, _ := rand.Int(rand.Reader, big.NewInt(int64(k)))
		sb.WriteByte(alphabet[c.Int64()])
	}

	return sb.String()
}

// RandomOwner generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generates a random currency code
func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "CAD"}
	n := len(currencies)
	index, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return currencies[index.Int64()]
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
