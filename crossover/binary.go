package crossover

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
)

type binary struct{}

func (binary) IsCompatible(chromosomeType reflect.Kind, flags uint) bool {
	return chromosomeType == reflect.Int // TODO no permutation should be allowed
}

type kPoints struct {
	binary

	k      int
	xps    []int
	buffer []int
}

func KPoints(k int) Operator {
	if k <= 0 {
		panic(fmt.Sprintf("invalid k-point crossover: k= %d", k)) // TODO
	}
	return &kPoints{k: k, xps: make([]int, k)}
}

func SinglePoint() Operator {
	return KPoints(1)
}

func TwoPoints() Operator {
	return KPoints(2)
}

func (s *kPoints) Crossover(mom, dad, child1, child2 []byte) error {
	totBits := len(mom) * 8
	if s.k >= totBits {
		return fmt.Errorf("cannot apply %d-point crossover to chromosomes %d bits long", s.k, totBits)
	}

	copy(child1, mom)
	copy(child2, dad)

	xps := s.randomXPoints(totBits)

	prevXByte := -1

	for i, xp := range xps {
		xByte := xp / 8
		xBits := xp % 8
		mask := ^(byte(0xff) << xBits)

		if i%2 == 1 {
			if xByte == prevXByte {
				// rollback crossover to high bits of current byte
				flip(child1, child2, xByte, ^mask)
				goto next
			}

			// apply crossover to previous slice
			copy(child1[prevXByte+1:xByte], dad[prevXByte+1:xByte])
			copy(child2[prevXByte+1:xByte], mom[prevXByte+1:xByte])

		} else {
			// apply crossover to high bits...
			mask = ^mask
		}
		// will apply crossover to current byte
		flip(child1, child2, xByte, mask)

	next:
		prevXByte = xByte
	}

	if s.k%2 == 1 {
		copy(child1[prevXByte+1:], dad[prevXByte+1:])
		copy(child2[prevXByte+1:], mom[prevXByte+1:])
	}

	return nil
}

func (s *kPoints) randomXPoints(totBits int) []int {
	xpRange := totBits - 2
	if len(s.buffer) == xpRange {
		// no need to change anything
	} else {
		if len(s.buffer) < xpRange {
			s.buffer = make([]int, xpRange)
		}
		for i := range s.buffer[:xpRange] {
			s.buffer[i] = 1 + i
		}
	}
	rand.Shuffle(xpRange, s.swapBuffer)

	copy(s.xps, s.buffer)
	slices.Sort(s.xps)
	return s.xps
}

func (s kPoints) swapBuffer(i, j int) {
	s.buffer[i], s.buffer[j] = s.buffer[j], s.buffer[i]
}

type uniform struct {
	binary

	rate float64
}

// TODO uniform evolving children separately

// TODO coarser grained option? k-size blocks
func ParametricHalfUniform(rate float64) Operator {
	return uniform{rate: rate}
}

func (u uniform) Crossover(mom, dad, child1, child2 []byte) error {
	copy(child1, mom)
	copy(child2, dad)

	for i := range len(mom) {
		mask := child1[i] ^ child2[i]
		for j := range 8 {
			m := (byte(1) << j)
			if mask&m != 0 && rand.Float64() >= u.rate {
				mask &^= m
			}
		}

		flip(child1, child2, i, mask)
	}
	return nil
}

func flip(a, b []byte, i int, mask byte) {
	av := a[i]
	bv := b[i]

	a[i] &^= mask
	a[i] |= bv & mask

	b[i] &^= mask
	b[i] |= av & mask
}
