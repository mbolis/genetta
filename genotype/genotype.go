package genotype

import (
	"fmt"
	"reflect"
	"unsafe"
)

type Schema[T any] struct {
	chromosomes []Chromosome
	sizeInBytes int
}

func New[T any](chromosomes ...Chromosome) (s Schema[T]) {
	s.chromosomes = chromosomes
	for _, c := range chromosomes {
		c.bytesIndex = s.sizeInBytes
		s.sizeInBytes += c.bytesLength
	}
	return
}

func (s Schema[T]) Size() int {
	return s.sizeInBytes
}

func (s Schema[T]) Bounds(i int) (start, end int) {
	start = i * s.sizeInBytes
	end = start + s.sizeInBytes
	return
}

func (s Schema[T]) Make(popSize int) []byte {
	return make([]byte, popSize*s.sizeInBytes)
}

func (s Schema[T]) Init() (t T) {
	if any(t) != nil {
		return
	}

	tt := reflect.TypeFor[T]()
	switch tt.Kind() {
	case reflect.Slice:
		cells := -1
		for _, c := range s.chromosomes {
			for _, g := range c.genes {
				if g.byteIndex > cells {
					cells = g.byteIndex
				}
			}
		}
		cells++
		// FIXME broken for complex numbers!
		t = reflect.MakeSlice(tt, cells, cells).Interface().(T)

	default:
		panic(fmt.Sprintf("unrecognized type: %s", tt))
	}
	return
}

func (s Schema[T]) Encode(ptr *T, data []byte) {
	for _, c := range s.chromosomes {
		for _, gene := range c.genes {
			gene.Encode(unsafe.Pointer(ptr), data[c.bytesIndex:])
		}
	}
}
func (s Schema[T]) Decode(ptr *T, data []byte) {
	for _, c := range s.chromosomes {
		for _, gene := range c.genes {
			gene.Decode(unsafe.Pointer(ptr), data[c.bytesIndex:])
		}
	}
}

func (s Schema[T]) Randomize(data []byte) {
	for _, c := range s.chromosomes {
		c.Randomize(data)
	}
}

func (s Schema[T]) Crossover(mom, dad, child1, child2 []byte) error {
	for _, c := range s.chromosomes {
		i1 := c.bytesIndex
		i2 := i1 + c.bytesLength

		err := c.crossover.Crossover(
			mom[i1:i2], dad[i1:i2],
			child1[i1:i2], child2[i1:i2],
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Schema[T]) Mutate(genotypes ...[]byte) error {
	for _, g := range genotypes {
		for _, c := range s.chromosomes {
			err := c.mutate.Mutate(g[c.bytesIndex : c.bytesIndex+c.bytesLength])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
