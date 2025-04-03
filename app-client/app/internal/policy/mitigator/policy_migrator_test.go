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

func BenchmarkSolveChallenge_5zeros(b *testing.B) {
	benchmarkSolveChallenge(b, 5)
}

func BenchmarkSolveChallenge_10zeros(b *testing.B) {
	benchmarkSolveChallenge(b, 10)
}

func BenchmarkSolveChallenge_15zeros(b *testing.B) {
	benchmarkSolveChallenge(b, 15)
}

func BenchmarkSolveChallenge_20zeros(b *testing.B) {
	benchmarkSolveChallenge(b, 20)
}

func BenchmarkSolveChallenge_25zeros(b *testing.B) {
	benchmarkSolveChallenge(b, 25)
}

func BenchmarkSolveChallenge_30zeros(b *testing.B) {
	benchmarkSolveChallenge(b, 30)
}

func benchmarkSolveChallenge(b *testing.B, difficulty int32) {
	// Генерируем 32 байта случайных данных.
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("Failed to generate random bytes: %v", err)
	}

	// Формируем буфер из 8 байт метки времени + 32 байта случайных данных.
	buf := make([]byte, 8+len(randomBytes))
	binary.BigEndian.PutUint64(buf[0:8], uint64(time.Now().Unix()))
	copy(buf[8:], randomBytes)

	// Вычисляем хэш от сформированного буфера.
	hash := sha256.Sum256(buf)

	// Отключаем вывод логов в бенчмарке.
	log.SetOutput(io.Discard)

	// Создаём решатель с большим временем ожидания.
	solver := NewPoWSolver(time.Hour)
	challenge := PoWChallenge{
		Timestamp:   time.Now().Unix(),
		RandomBytes: hash[:],
		Difficulty:  difficulty,
	}

	ctx := context.Background()

	// Сбросим таймер бенчмарка после подготовки.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, solverErr := solver.SolvePoWChallenge(ctx, challenge)
		if solverErr != nil {
			b.Fatalf("Error solving PoW challenge: %v", solverErr)
		}
	}
}
