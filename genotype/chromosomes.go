package genotype

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"unsafe"

	"github.com/mbolis/genetta/crossover"
	"github.com/mbolis/genetta/mutation"
)

type Chromosome struct {
	type_ reflect.Kind
	flags Flags

	genes       []Gene
	bytesLength int
	bytesIndex  int

	crossover crossover.Operator
	mutate    mutation.Operator
}

type Flags uint

const (
	FlagDecimal Flags = 1 << iota
	FlagPermutation
)

func IntChromosome(components ...ChromosomeComponent) (c Chromosome) {
	c.type_ = reflect.Int
	for _, comp := range components {
		comp.apply(&c)
	}
	return
}

type ChromosomeComponent interface {
	apply(*Chromosome)
}

func (c Chromosome) Randomize(data []byte) {
	switch c.type_ {
	case reflect.Int:
		for i := range data {
			data[i] = byte(rand.Uint())
			// TODO enforce min-max
		}
	case reflect.Float64:
		for i := range data {
			*(*float64)(unsafe.Pointer(&data[i])) = rand.Float64()
			// TODO enforce min-max
		}
	case reflect.Float32:
		for i := range data {
			*(*float32)(unsafe.Pointer(&data[i])) = rand.Float32()
			// TODO enforce min-max
		}
	default:
		panic(fmt.Sprintf("invalid chromosome type: %d", c.type_))
	}
}
