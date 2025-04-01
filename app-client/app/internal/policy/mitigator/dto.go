package mitigator

// PoWChallenge represents the data structure for the Proof-of-Work challenge
// received from the server (using JSON).
type PoWChallenge struct {
	Timestamp   int64  `json:"timestamp"`
	RandomBytes []byte `json:"random_bytes"` // Ensure server sends bytes correctly encoded in JSON (e.g., base64)
	Difficulty  int32  `json:"difficulty"`
}

// PoWSolution represents the data structure for the Proof-of-Work solution
// sent to the server (using JSON).
type PoWSolution struct {
	Nonce uint64 `json:"nonce"`
}

// QuoteResponse represents the data structure for the quote received from the server.
type QuoteResponse struct {
	Quote string `json:"quote"`
	Error string `json:"error,omitempty"` // Optional: if server can send errors this way
}
