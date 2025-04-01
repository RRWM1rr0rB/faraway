package tcp

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// PoWChallenge struct.
type PoWChallenge struct {
	Timestamp   int64
	RandomBytes []byte
	Difficulty  int32
}

// PoWSolution struct.
type PoWSolution struct {
	Nonce uint64
}

// GeneratePoWChallenge create new PoW-task.
func GeneratePoWChallenge(difficulty int32) (*PoWChallenge, error) {
	if difficulty < 0 || difficulty > 256 {
		return nil, fmt.Errorf("invalid difficulty")
	}

	randomBytes := make([]byte, 32)

	if _, err := cryptorand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return &PoWChallenge{
		Timestamp:   time.Now().Unix(),
		RandomBytes: randomBytes,
		Difficulty:  difficulty,
	}, nil
}

// ValidatePoWSolution check solution  PoW-task.
func ValidatePoWSolution(challenge *PoWChallenge, solution *PoWSolution) bool {
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

// WritePoWChallenge write PoW-задачу в writer.
func WritePoWChallenge(w io.Writer, challenge *PoWChallenge) error {
	// Write timestamp (8 byte)
	if err := binary.Write(w, binary.BigEndian, challenge.Timestamp); err != nil {
		return fmt.Errorf("failed to write timestamp: %w", err)
	}

	// Write random bytes (32 byte)
	if _, err := w.Write(challenge.RandomBytes); err != nil {
		return fmt.Errorf("failed to write random bytes: %w", err)
	}

	// Write difficulty (4 byte)
	if err := binary.Write(w, binary.BigEndian, challenge.Difficulty); err != nil {
		return fmt.Errorf("failed to write difficulty: %w", err)
	}

	return nil
}

// ReadPoWChallenge read PoW-task from reader.
func ReadPoWChallenge(r io.Reader) (*PoWChallenge, error) {
	challenge := &PoWChallenge{
		RandomBytes: make([]byte, 32),
	}

	// Read timestamp (8 byte)
	if err := binary.Read(r, binary.BigEndian, &challenge.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}

	// Read random bytes (32 byte)
	if _, err := io.ReadFull(r, challenge.RandomBytes); err != nil {
		return nil, fmt.Errorf("failed to read random bytes: %w", err)
	}

	// Читаем difficulty (4 byte)
	if err := binary.Read(r, binary.BigEndian, &challenge.Difficulty); err != nil {
		return nil, fmt.Errorf("failed to read difficulty: %w", err)
	}

	return challenge, nil
}

// WritePoWSolution write solution PoW-task to writer.
func WritePoWSolution(w io.Writer, solution *PoWSolution) error {
	return binary.Write(w, binary.BigEndian, solution.Nonce)
}

// ReadPoWSolution read solution PoW-task from reader.
func ReadPoWSolution(r io.Reader) (*PoWSolution, error) {
	var solution PoWSolution
	if err := binary.Read(r, binary.BigEndian, &solution.Nonce); err != nil {
		return nil, fmt.Errorf("failed to read nonce: %w", err)
	}
	return &solution, nil
}
