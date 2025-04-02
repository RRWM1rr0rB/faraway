package mitigator

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
)

// GetWisdom returns a random piece of wisdom.
func (p *StaticWisdomProvider) GetWisdom(ctx context.Context) (WisdomDTO, error) {
	if len(p.quotes) == 0 {
		logging.L(ctx).Warn("No quotes configured")
		return WisdomDTO{}, ErrNoWisdomFound
	}

	// Select a random quote
	index := p.r.Intn(len(p.quotes))
	quote := p.quotes[index]

	logging.L(ctx).Info("Providing wisdom", logging.StringAttr("quote", quote))

	return WisdomDTO{Quote: quote}, nil
}

// GeneratePoWChallenge create new PoW-task.
func (p *StaticWisdomProvider) GeneratePoWChallenge(difficulty int32) (*PoWChallenge, error) {
	if difficulty < 0 || difficulty > 256 {
		return nil, fmt.Errorf("invalid difficulty")
	}

	randomBytes := make([]byte, 32)

	if _, err := cryptorand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	buf := make([]byte, 8+len(randomBytes))
	binary.BigEndian.PutUint64(buf[0:8], uint64(time.Now().Unix()))
	copy(buf[8:], randomBytes)

	hash := sha256.Sum256(buf)

	return &PoWChallenge{
		Timestamp:   time.Now().Unix(),
		RandomBytes: hash[:],
		Difficulty:  difficulty,
	}, nil
}

// ValidatePoWSolution check solution  PoW-task.
func (p *StaticWisdomProvider) ValidatePoWSolution(challenge *PoWChallenge, solution *PoWSolution) bool {
	if time.Now().Unix()-challenge.Timestamp > 60 {
		return false
	}

	buf := make([]byte, 8+32+8)
	binary.BigEndian.PutUint64(buf[0:8], uint64(challenge.Timestamp))
	copy(buf[8:40], challenge.RandomBytes)
	binary.BigEndian.PutUint64(buf[40:48], solution.Nonce)

	hash := sha256.Sum256(buf)

	leadingZeros := countLeadingZeros(hash[:])
	return leadingZeros >= challenge.Difficulty
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
