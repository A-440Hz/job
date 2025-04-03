package db

import "crypto/rand"

type IdPrefix string

const (
	UserIdPrefix    IdPrefix = "usr"
	TrackerIdPrefix IdPrefix = "tkr"
	ItemIdPrefix    IdPrefix = "itm"

	DefaultSuffixLen int = 8
	b62Chars             = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func makeIdSuffix(ln int) string {
	byteSlice := make([]byte, ln)

	// this populates byteSlice with random bytes
	// crypto/rand gets randomness from the OS's CSPRNG (/dev/random for linux)
	// high entropy and safe with concurrency, but slower than math/rand
	rand.Read(byteSlice)
	for i, b := range byteSlice {
		byteSlice[i] = b62Chars[b%byte(len(b62Chars))]
	}
	return string(byteSlice)
}

func NewID(prefix IdPrefix) string {
	return string(prefix) + "_" + makeIdSuffix(DefaultSuffixLen)
}
