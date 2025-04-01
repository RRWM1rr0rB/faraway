package mitigator

import "log"

func NewChallenger(numberLeadingZeros, challengeNumberSymbols int32, solutionNumberSymbols int32) *Challenger {
	return &Challenger{
		numberLeadingZeros:     numberLeadingZeros,
		challengeNumberSymbols: challengeNumberSymbols,
		solutionNumberSymbols:  solutionNumberSymbols,
	}
}

func (c *Challenger) GetChallenge() ChallengeInfo {
	return ChallengeInfo{
		RandomString:          generateCryptoRandomString(c.challengeNumberSymbols),
		NumberLeadingZeros:    c.numberLeadingZeros,
		SolutionNumberSymbols: c.solutionNumberSymbols,
	}
}

func (c *Challenger) ValidateSolution(challenge ChallengeInfo, solution string) bool {
	// all parameters get from challenge not from challenger because we can use range for zeros and length
	if len(solution) != int(challenge.SolutionNumberSymbols) {
		log.Printf("wrong number symbols in solution: %v, want: %v", len(solution), challenge.SolutionNumberSymbols)

		return false
	}

	return isValid(challenge.RandomString+solution, challenge.NumberLeadingZeros)
}
