package mutation

import "reflect"

type Operator interface {
	Mutate(genotype []byte) error
	IsCompatible(chromosomeType reflect.Kind, flags uint) bool
}
