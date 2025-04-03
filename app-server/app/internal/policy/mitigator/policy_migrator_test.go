package mitigator_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"app-server/app/internal/policy/mitigator"
)

func TestGetWisdom_EmptyQuotes(t *testing.T) {
	provider := &mitigator.StaticWisdomProvider{
		Quotes: []string{},
		R:      rand.New(rand.NewSource(1)),
	}
	ctx := context.Background()

	_, err := provider.GetWisdom(ctx)
	assert.ErrorIs(t, err, mitigator.ErrNoWisdomFound)
}

func TestGeneratePoWChallenge_ValidDifficulty(t *testing.T) {
	provider := mitigator.New()

	tests := []struct {
		name       string
		difficulty int32
	}{
		{"Min", 0},
		{"Low", 1},
		{"High", 255},
		{"Max", 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge, err := provider.GeneratePoWChallenge(tt.difficulty)
			require.NoError(t, err)
			assert.Equal(t, tt.difficulty, challenge.Difficulty)
			assert.Len(t, challenge.RandomBytes, 32)
			assert.True(t, challenge.Timestamp <= time.Now().Unix())
		})
	}
}

func TestGeneratePoWChallenge_InvalidDifficulty(t *testing.T) {
	provider := mitigator.New()

	tests := []struct {
		name       string
		difficulty int32
	}{
		{"Negative", -1},
		{"TooHigh", 257},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.GeneratePoWChallenge(tt.difficulty)
			assert.Error(t, err)
		})
	}
}

func TestValidatePoWSolution_Valid(t *testing.T) {
	provider := mitigator.New()
	challenge, err := provider.GeneratePoWChallenge(1)
	require.NoError(t, err)

	var solution *mitigator.PoWSolution
	for nonce := uint64(0); nonce < 100000; nonce++ {
		sol := &mitigator.PoWSolution{Nonce: nonce}
		if provider.ValidatePoWSolution(challenge, sol) {
			solution = sol
			break
		}
	}

	require.NotNil(t, solution, "should find valid solution")
	assert.True(t, provider.ValidatePoWSolution(challenge, solution))
}

func TestValidatePoWSolution_Expired(t *testing.T) {
	provider := mitigator.New()
	challenge := &mitigator.PoWChallenge{
		Timestamp:   time.Now().Unix() - 61,
		RandomBytes: make([]byte, 32),
		Difficulty:  1,
	}
	solution := &mitigator.PoWSolution{Nonce: 0}

	assert.False(t, provider.ValidatePoWSolution(challenge, solution))
}

func TestValidatePoWSolution_InvalidNonce(t *testing.T) {
	provider := mitigator.New()
	challenge, err := provider.GeneratePoWChallenge(8)
	require.NoError(t, err)

	solution := &mitigator.PoWSolution{Nonce: 12345}
	assert.False(t, provider.ValidatePoWSolution(challenge, solution))
}

func TestCountLeadingZeros(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int32
	}{
		{
			name:     "AllZeros",
			data:     make([]byte, 32),
			expected: 256,
		},
		{
			name:     "FirstByteZeroSecond0x80",
			data:     []byte{0x00, 0x80},
			expected: 8,
		},
		{
			name:     "SingleByteLeadingZeros",
			data:     []byte{0x01},
			expected: 7,
		},
		{
			name:     "NoLeadingZeros",
			data:     []byte{0xFF, 0xFF},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mitigator.CountLeadingZeros(tt.data)
			assert.Equal(t, tt.expected, got)
		})
	}
}
