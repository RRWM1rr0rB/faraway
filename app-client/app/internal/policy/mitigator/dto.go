package mitigator

type PoWChallenge struct {
	Timestamp   int64
	RandomBytes []byte
	Difficulty  int32
}

type PoWSolution struct {
	Nonce uint64
}
type Quote struct {
	Quote string
}
