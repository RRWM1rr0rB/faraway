package mitigator

import (
	"math/rand"
	"time"
)

// StaticWisdomProvider provides wisdom from a predefined list.
type StaticWisdomProvider struct {
	Quotes []string
	R      *rand.Rand
}

// New creates a new StaticWisdomProvider.
func New() *StaticWisdomProvider {
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
		Quotes: quotes,
		R:      randomGenerator,
	}
}
