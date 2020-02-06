package manager

import (
	"math/rand"
	"strings"
	"time"
)

// JoinTokens : defines a managers join tokens
type JoinTokens struct {
	Manager string
	Worker  string
}

// GenerateToken : generates a random token for cluster nodes to use
func GenerateToken() string {
	rand.Seed(time.Now().UnixNano())

	chars := []rune(
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			"0123456789",
	)

	length := 50

	var t strings.Builder

	for i := 0; i < length; i++ {
		t.WriteRune(chars[rand.Intn(len(chars))])
	}

	return t.String()
}
