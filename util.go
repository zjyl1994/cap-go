package cap

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
)

func prng(seed string, length int) string {
	h := fnv.New32a()
	h.Write([]byte(seed))
	state := h.Sum32()
	result := ""

	next := func() uint32 {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		return state
	}

	for len(result) < length {
		rnd := next()
		hexStr := fmt.Sprintf("%08x", rnd)
		result += hexStr
	}

	return result[:length]
}

func sha256Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}
