package mitigator

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

// StaticWisdomProvider provides wisdom from a predefined list.
type StaticWisdomProvider struct {
	quotes []string
	log    *slog.Logger
	r      *rand.Rand // Source for random numbers
}

// New creates a new StaticWisdomProvider.
func New(log *slog.Logger) *StaticWisdomProvider {
	quotes := []string{
		"The greatest glory in living lies not in never falling, but in rising every time we fall. - Nelson Mandela",
		"The way to get started is to quit talking and begin doing. - Walt Disney",
		"Your time is limited, don't waste it living someone else's life. - Steve Jobs",
		"If life were predictable it would cease to be life, and be without flavor. - Eleanor Roosevelt",
		"If you look at what you have in life, you'll always have more. If you look at what you don't have in life, you'll never have enough. - Oprah Winfrey",
		"Life is what happens when you're busy making other plans. - John Lennon",
		"Spread love everywhere you go. Let no one ever come to you without leaving happier. - Mother Teresa",
		"Tell me and I forget. Teach me and I remember. Involve me and I learn. - Benjamin Franklin",
		"The best and most beautiful things in the world cannot be seen or even touched - they must be felt with the heart. - Helen Keller",
		"It is during our darkest moments that we must focus to see the light. - Aristotle",
	}

	// Seed the random number generator
	source := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(source)

	return &StaticWisdomProvider{
		quotes: quotes,
		log:    log.With(slog.String("component", "wisdom_provider")),
		r:      randomGenerator,
	}
}

// GetWisdom returns a random piece of wisdom.
func (p *StaticWisdomProvider) GetWisdom(ctx context.Context) (WisdomDTO, error) {
	if len(p.quotes) == 0 {
		p.log.WarnContext(ctx, "No quotes configured")
		return WisdomDTO{}, ErrNoWisdomFound
	}

	// Select a random quote
	index := p.r.Intn(len(p.quotes))
	quote := p.quotes[index]

	p.log.DebugContext(ctx, "Providing wisdom", slog.String("quote", quote))

	return WisdomDTO{Quote: quote}, nil
}
