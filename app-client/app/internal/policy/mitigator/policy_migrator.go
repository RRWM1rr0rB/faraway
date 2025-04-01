package mitigator

import (
	"fmt"
	"log"
	"time"
)

type Solver struct {
	timeout time.Duration
}

func NewSolver(timeout time.Duration) *Solver {
	return &Solver{
		timeout: timeout,
	}
}

func (s *Solver) SolveChallenge(task PoWCalc) (string, error) {
	timer := time.NewTimer(s.timeout)
	defer timer.Stop()

	var (
		numberAttempts int
		solution       string
	)

	for {
		select {
		case <-timer.C:
			return "", fmt.Errorf("timeout: (%v)", s.timeout)
		default:
		}

		solution = generateMathRandomString(task.SolutionNumberSymbols)
		numberAttempts++

		if isValid(task.RandomString+solution, task.NumberLeadingZeros) {
			log.Printf("solution %v found after %v attempts", solution, numberAttempts)

			break
		}
	}

	return solution, nil
}
