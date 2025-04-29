package crossover

import (
	"math/rand/v2"
	"reflect"
)

type Operator interface {
	Crossover(mom, dad, child1, child2 []byte) error
	IsCompatible(chromosomeType reflect.Kind, flags uint) bool
}

type probability struct {
	probability float64
	Operator
}

func Probability(p float64, s Operator) Operator {
	return probability{p, s}
}

func (p probability) Crossover(mom, dad, child1, child2 []byte) error {
	if rand.Float64() >= p.probability {
		copy(child1, mom)
		copy(child2, dad)
		return nil
	}

	return p.Operator.Crossover(mom, dad, child1, child2)
}
