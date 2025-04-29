package mutation

import (
	"math/rand/v2"
	"reflect"
)

type binary struct{}

func (binary) IsCompatible(chromosomeType reflect.Kind, flags uint) bool {
	return chromosomeType == reflect.Int // TODO no permutation should be allowed
}

type bitString struct {
	binary
	n float64
}

func BitString(meanFlips int) Operator {
	return bitString{n: float64(meanFlips)}
}

func (b bitString) Mutate(genome []byte) error {
	p := b.n / float64(len(genome)*8) // XXX cache? XXX Not exact!

	for i, b := range genome { // TODO maybe convert to []uint64 with unsafe?
		var mask byte
		for j := range 8 {
			if rand.Float64() < p {
				mask |= 1 << j
			}
		}
		if mask == 0 {
			continue
		}

		genome[i] = b ^ mask
	}
	return nil
}
