package mitigator

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

func SolvePoWChallenge(challenge *PoWChallenge) (*PoWSolution, error) {
	// Search for a solution.
	start := time.Now()
	var nonce uint64

	buf := make([]byte, 8+32+8)
	// Write timestamp
	binary.BigEndian.PutUint64(buf[0:8], uint64(challenge.Timestamp))
	// Copy random bytes.
	copy(buf[8:40], challenge.RandomBytes)

	for nonce = 0; nonce < math.MaxUint64; nonce++ {
		// Write nonce.
		binary.BigEndian.PutUint64(buf[40:48], nonce)
		hash := sha256.Sum256(buf)
		if countLeadingZeros(hash[:]) >= challenge.Difficulty {
			return &PoWSolution{Nonce: nonce}, nil
		}
		// If search took more than 60 seconds, return error.
		if time.Since(start) > 60*time.Second {
			return nil, fmt.Errorf("timeout while solving PoW challenge")
		}
	}

	return nil, fmt.Errorf("solution not found")
}

// countLeadingZeros counts the number of leading zeros in a byte slice.
func countLeadingZeros(data []byte) int32 {
	var zeros int32
	for _, b := range data {
		if b == 0 {
			zeros += 8
		} else {
			// Count the number of leading zeros.
			for i := 7; i >= 0; i-- {
				if (b >> i) == 0 {
					zeros++
				} else {
					return zeros
				}
			}
		}
	}
	return zeros
}
