package tcp

import (
	"log"
	"net"
	"sync"
	"time"
)

const (
	defaultRateLimitPerIP    = 10             // Max connection per sec.
	defaultInitialDifficulty = 4              // Star difficulty PoW(0000).
	maxDifficulty            = 8              // Max Difficulty PoW.
	banDuration              = 24 * time.Hour // Ban Duration.
)

type RateLimiter struct {
	mu sync.RWMutex
	// IP -> amount connection in this is second.
	connectionsPerIP map[string]*rateCounter
	// IP -> ban end time.
	bannedIPs map[string]time.Time
	// IP -> currently difficulty PoW.
	difficulties map[string]int32
	logger       *log.Logger
}

type rateCounter struct {
	count     int64
	timestamp time.Time
}

// NewRateLimiter create new RateLimiter.
func NewRateLimiter(logger *log.Logger) *RateLimiter {
	if logger == nil {
		logger = log.Default()
	}
	return &RateLimiter{
		connectionsPerIP: make(map[string]*rateCounter),
		bannedIPs:        make(map[string]time.Time),
		difficulties:     make(map[string]int32),
		logger:           logger,
	}
}

// RateLimitMiddleware creates a middleware for limiting connections
//
// This middleware checks if the client's IP is banned, and if the rate limit
// for the IP is exceeded. If the rate limit is exceeded, the client is sent a
// PoW challenge with increasing difficulty.
func RateLimitMiddleware(limiter *RateLimiter) func(net.Conn) {
	return func(conn net.Conn) {
		ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()

		// Check if the IP is banned
		if limiter.isBanned(ip) {
			limiter.logger.Printf("IP %s is banned", ip)
			err := conn.Close()
			if err != nil {
				log.Printf("failed to close connection: %v", err)
			}
			return
		}

		// Check if the rate limit is exceeded
		if !limiter.allowConnection(ip) {
			limiter.logger.Printf("Rate limit exceeded for IP %s", ip)
			// Increase the difficulty of the PoW challenge
			limiter.increaseDifficulty(ip)
			err := conn.Close()
			if err != nil {
				log.Printf("failed to close connection: %v", err)
			}
			return
		}

		// Get the current difficulty for the IP
		difficulty := limiter.getDifficulty(ip)

		// Generate a PoW challenge
		challenge, err := GeneratePoWChallenge(difficulty)
		if err != nil {
			limiter.logger.Printf("Failed to generate PoW challenge: %v", err)
			connErr := conn.Close()
			if connErr != nil {
				log.Printf("failed to close connection: %v", connErr)
			}
			return
		}

		// Send the challenge to the client
		if writePoWChallengeErr := WritePoWChallenge(conn, challenge); writePoWChallengeErr != nil {
			limiter.logger.Printf("Failed to write PoW challenge: %v", writePoWChallengeErr)
			connErr := conn.Close()
			if connErr != nil {
				log.Printf("failed to close connection: %v", connErr)
			}
			return
		}

		// Read the solution from the client
		solution, solutionErr := ReadPoWSolution(conn)
		if solutionErr != nil {
			limiter.logger.Printf("Failed to read PoW solution: %v", err)
			connErr := conn.Close()
			if connErr != nil {
				log.Printf("failed to close connection: %v", connErr)
			}
			return
		}

		// Validate the solution
		if !ValidatePoWSolution(challenge, solution) {
			limiter.logger.Printf("Invalid PoW solution from IP %s", ip)
			limiter.increaseDifficulty(ip)
			connErr := conn.Close()
			if connErr != nil {
				log.Printf("failed to close connection: %v", connErr)
			}
			return
		}

		// If the solution is valid, decrease the difficulty
		limiter.decreaseDifficulty(ip)
		limiter.logger.Printf("Valid PoW solution from IP %s, difficulty: %d", ip, difficulty)
	}
}

func (r *RateLimiter) isBanned(ip string) bool {
	r.mu.RLock()
	banTime, exists := r.bannedIPs[ip]
	if !exists {
		r.mu.RUnlock()
		return false
	}

	if time.Now().After(banTime) {
		r.mu.RUnlock()
		r.mu.Lock()
		delete(r.bannedIPs, ip)
		r.mu.Unlock()
		return false
	}
	r.mu.RUnlock()
	return true
}

// allowConnection проверяет, разрешено ли новое соединение
func (r *RateLimiter) allowConnection(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	counter, exists := r.connectionsPerIP[ip]

	if !exists {
		r.connectionsPerIP[ip] = &rateCounter{
			count:     1,
			timestamp: now,
		}
		return true
	}

	// Если прошла секунда, сбрасываем счетчик
	if now.Sub(counter.timestamp) >= time.Second {
		counter.count = 1
		counter.timestamp = now
		return true
	}

	// Увеличиваем счетчик
	counter.count++

	// Если превышен лимит, бан IP
	if counter.count > defaultRateLimitPerIP {
		r.bannedIPs[ip] = now.Add(banDuration)
		return false
	}

	return true
}

// increaseDifficulty увеличивает сложность PoW для IP
func (r *RateLimiter) increaseDifficulty(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current := r.difficulties[ip]
	if current < maxDifficulty {
		r.difficulties[ip] = current + 1
		r.logger.Printf("Increased PoW difficulty for IP %s to %d", ip, current+1)
	}
}

// decreaseDifficulty уменьшает сложность PoW для IP
func (r *RateLimiter) decreaseDifficulty(ip string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current := r.difficulties[ip]
	if current > defaultInitialDifficulty {
		r.difficulties[ip] = current - 1
		r.logger.Printf("Decreased PoW difficulty for IP %s to %d", ip, current-1)
	}
}

// getDifficulty возвращает текущую сложность PoW для IP
func (r *RateLimiter) getDifficulty(ip string) int32 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	difficulty, exists := r.difficulties[ip]
	if !exists {
		return defaultInitialDifficulty
	}
	return difficulty
}

// Cleanup удаляет устаревшие записи
func (r *RateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Очищаем истекшие баны
	for ip, banTime := range r.bannedIPs {
		if now.After(banTime) {
			delete(r.bannedIPs, ip)
		}
	}

	// Очищаем старые счетчики
	for ip, counter := range r.connectionsPerIP {
		if now.Sub(counter.timestamp) >= time.Second {
			delete(r.connectionsPerIP, ip)
		}
	}
}
