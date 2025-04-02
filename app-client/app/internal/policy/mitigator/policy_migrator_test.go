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

// benchmarkSolveChallenge теперь будет хешировать случайные байты при создании задачи.
func BenchmarkSolveChallenge_5zeros_5symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 5, 5)
}

func BenchmarkSolveChallenge_5zeros_20symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 5, 20)
}

func BenchmarkSolveChallenge_10zeros_5symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 10, 5)
}

func BenchmarkSolveChallenge_10zeros_20symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 10, 20)
}

func BenchmarkSolveChallenge_15zeros_5symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 15, 5)
}

func BenchmarkSolveChallenge_15zeros_20symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 15, 20)
}

func BenchmarkSolveChallenge_20zeros_5symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 20, 5)
}

func BenchmarkSolveChallenge_20zeros_20symbols(b *testing.B) {
	benchmarkSolveChallenge(b, 20, 20)
}

// benchmarkSolveChallenge теперь генерирует challenge с хешированием
func benchmarkSolveChallenge(b *testing.B, difficulty int32, numSymbols int) {
	// Генерируем случайные байты для challenge
	randomBytes := make([]byte, numSymbols)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("Failed to generate random bytes: %v", err)
	}

	// Хешируем случайные байты вместе с временной меткой
	buf := make([]byte, 8+len(randomBytes)) // Временная метка + случайные байты
	binary.BigEndian.PutUint64(buf[0:8], uint64(time.Now().Unix()))
	copy(buf[8:], randomBytes)

	hash := sha256.Sum256(buf) // Хешируем

	// Логирование отключается для бенчмарков
	log.SetOutput(io.Discard)

	// Создаем challenge с хешированными байтами
	solver := NewPoWSolver(time.Hour)
	challenge := PoWChallenge{
		Timestamp:   time.Now().Unix(),
		RandomBytes: hash[:], // Используем хешированные байты
		Difficulty:  difficulty,
	}

	// Контекст для бенчмарка
	ctx := context.Background()

	// Выполнение бенчмарка N раз
	for i := 0; i < b.N; i++ {
		_, err := solver.SolvePoWChallenge(ctx, challenge)
		if err != nil {
			b.Fatalf("Error solving PoW challenge: %v", err)
		}
	}
}
