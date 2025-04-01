package mitigator

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"
)

// SolvePoWChallenge attempts to find a valid nonce for the given challenge.
func (p *PoWSolver) SolvePoWChallenge(ctx context.Context, challenge PoWChallenge) (*PoWSolution, error) {
	logger := logging.L(ctx).With(logging.StringAttr("component", "pow_solver"))
	logger.Info("Attempting to solve PoW challenge",
		logging.Int64Attr("timestamp", challenge.Timestamp),
		logging.IntAttr("difficulty", int(challenge.Difficulty)))

	if challenge.Difficulty < 0 || challenge.Difficulty > 256 || len(challenge.RandomBytes) == 0 {
		logger.Error("Invalid challenge parameters received",
			logging.IntAttr("difficulty", int(challenge.Difficulty)),
			logging.IntAttr("bytes_len", len(challenge.RandomBytes)))
		return nil, ErrPoWInvalidInput
	}

	// Prepare constant part of the buffer for hashing
	prefixBuf := make([]byte, 8+len(challenge.RandomBytes)) // Timestamp (8) + RandomBytes
	binary.BigEndian.PutUint64(prefixBuf[0:8], uint64(challenge.Timestamp))
	copy(prefixBuf[8:], challenge.RandomBytes)

	// Create a context with timeout for solving
	solveCtx, cancel := context.WithTimeout(ctx, p.maxSolutionTime)
	defer cancel()

	var nonce uint64 = 0
	startTime := time.Now()

	for nonce < math.MaxUint64 {
		select {
		case <-solveCtx.Done(): // Check for timeout or parent context cancellation
			logger.Warn("PoW solving timed out or cancelled", logging.DurationAttr("elapsed", time.Since(startTime)))
			return nil, ErrPoWTimeout
		default:
			// Combine prefix and nonce for hashing
			hashBuf := make([]byte, len(prefixBuf)+8) // Prefix + Nonce (8)
			copy(hashBuf[0:], prefixBuf)
			binary.BigEndian.PutUint64(hashBuf[len(prefixBuf):], nonce)

			// Calculate hash
			hash := sha256.Sum256(hashBuf)

			// Check leading zeros
			if countLeadingZeros(hash[:]) >= challenge.Difficulty {
				elapsed := time.Since(startTime)
				logger.Info("PoW challenge solved",
					logging.Uint64Attr("nonce", nonce),
					logging.DurationAttr("elapsed", elapsed))
				return &PoWSolution{Nonce: nonce}, nil
			}
			nonce++

			// Optional: yield control occasionally to prevent blocking CPU entirely
			if nonce%1000000 == 0 { // Every million attempts
				runtime.Gosched()
			}
		}
	}

	// Should be practically unreachable with uint64 nonce space
	logger.Error("Failed to solve PoW challenge - nonce space exhausted")
	return nil, fmt.Errorf("failed to solve PoW challenge - nonce space exhausted")
}

// countLeadingZeros counts the number of leading zero bits in a byte slice.
func countLeadingZeros(data []byte) int32 {
	var zeros int32
	for _, b := range data {
		if b == 0 {
			zeros += 8
		} else {
			for i := 7; i >= 0; i-- {
				if (b>>i)&1 == 0 { // Check bit i
					zeros++
				} else {
					return zeros // Found the first '1' bit
				}
			}
		}
	}

	return zeros // Should not happen for non-zero SHA256 hash, but safe fallback.
}
