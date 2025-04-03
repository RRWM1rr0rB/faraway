package mitigator

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"log"
	"testing"
	"time"
)

func BenchmarkDifficulty_5zeros(b *testing.B) {
	benchmarkDifficulty(b, 5)
}

func BenchmarkDifficulty_10zeros(b *testing.B) {
	benchmarkDifficulty(b, 10)
}

func BenchmarkDifficulty_15zeros(b *testing.B) {
	benchmarkDifficulty(b, 15)
}

func BenchmarkDifficulty_20zeros(b *testing.B) {
	benchmarkDifficulty(b, 20)
}

func BenchmarkDifficulty_25zeros(b *testing.B) {
	benchmarkDifficulty(b, 25)
}

func BenchmarkDifficulty_30zeros(b *testing.B) {
	benchmarkDifficulty(b, 30)
}

func benchmarkDifficulty(b *testing.B, difficulty int32) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("Failed to generate random bytes: %v", err)
	}

	buf := make([]byte, 8+len(randomBytes))
	binary.BigEndian.PutUint64(buf[0:8], uint64(time.Now().Unix()))
	copy(buf[8:], randomBytes)

	hash := sha256.Sum256(buf)

	log.SetOutput(io.Discard)

	solver := NewPoWSolver(time.Hour)
	challenge := PoWChallenge{
		Timestamp:   time.Now().Unix(),
		RandomBytes: hash[:],
		Difficulty:  difficulty,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, solverErr := solver.SolvePoWChallenge(ctx, challenge)
		if solverErr != nil {
			b.Fatalf("Error solving PoW challenge: %v", solverErr)
		}
	}
}
