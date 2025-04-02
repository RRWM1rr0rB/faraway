package mitigator

// WisdomDTO holds the wisdom data.
type WisdomDTO struct {
	Quote string `json:"quote"`
}

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
