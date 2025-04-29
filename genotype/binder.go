package genotype

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"unsafe"

	"github.com/mbolis/genetta/crossover"
	"github.com/mbolis/genetta/mutation"
)

type BindFunc func(any) *GeneSpec

type binder[T any] struct {
	root  *T
	root_ T
	start uintptr
	end   uintptr
}

func newBinder[T any]() (b binder[T]) {
	t := reflect.TypeFor[T]()
	switch t.Kind() {
	case reflect.Slice:
		b.root_ = reflect.MakeSlice(t, 0, 0).Interface().(T)
	}

	b.root = &b.root_
	b.start = uintptr(unsafe.Pointer(b.root))
	b.end = b.start + unsafe.Sizeof(b.root_)
	return
}

func (b binder[T]) bind(position any) *GeneSpec {
	t := reflect.TypeOf(position)
	if t.Kind() != reflect.Pointer {
		panic("must use a pointer to bind") // TODO
	}
	t = t.Elem()

	v := reflect.ValueOf(position)
	ptr := v.Pointer()
	if ptr < b.start || ptr >= b.end {
		panic("pointer out of range") // TODO
	}

	return b.x(t, ptr)
}
func (b binder[T]) x(t reflect.Type, ptr uintptr) *GeneSpec {
	switch t.Kind() {
	case reflect.Array:
		if t.Len() == 0 {
			return &emptyGeneSpec
		}

		v := reflect.NewAt(t, unsafe.Pointer(ptr))
		g := b.bind(v.Elem().Index(0).Addr().Interface())
		g.len = t.Len()
		return g

	case reflect.Slice:
		g := b.x(t.Elem(), ptr)
		g.isSlice = true
		g.cells *= g.len
		g.len = 0
		return g

	case reflect.Bool:
		return &GeneSpec{
			type_:    t,
			bits:     1,
			cells:    1,
			len:      1,
			phOffset: ptr - b.start,
		}

	case
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &GeneSpec{
			type_:    t,
			bits:     t.Bits(),
			cells:    1,
			len:      1,
			phOffset: ptr - b.start,
		}

	case reflect.Complex64, reflect.Complex128:
		var ct reflect.Type
		switch t.Kind() {
		case reflect.Complex64:
			ct = reflect.TypeFor[float32]()
		case reflect.Complex128:
			ct = reflect.TypeFor[float64]()
		}

		return &GeneSpec{
			type_:    ct,
			bits:     t.Bits() / 2,
			cells:    2,
			len:      1,
			phOffset: ptr - b.start,
		}

	default:
		panic("invalid pointer type") // TODO
	}
}

func (b binder[T]) validate() error {
	if !memEq(b.root, &b.root_) {
		return errors.New("you should not change the root value") // TODO
	}
	return nil
}

type Spec []ChromosomeSpec

func (s *Spec) addChromosome(k reflect.Kind, genes ...*GeneSpec) *ChromosomeSpec {
	*s = append(*s, ChromosomeSpec{
		type_: k,
		genes: genes,
	})
	return &(*s)[len(*s)-1]
}
func (s *Spec) IntChromosome(genes ...*GeneSpec) *ChromosomeSpec {
	return s.addChromosome(reflect.Int, genes...).
		Crossover(crossover.Probability(0.75, crossover.SinglePoint()))
}
func (s *Spec) Float32Chromosome(genes ...*GeneSpec) *ChromosomeSpec {
	return s.addChromosome(reflect.Float32, genes...)
}
func (s *Spec) Float64Chromosome(genes ...*GeneSpec) *ChromosomeSpec {
	return s.addChromosome(reflect.Float64, genes...)
}

type ChromosomeSpec struct {
	type_     reflect.Kind
	flags     Flags
	genes     []*GeneSpec
	crossover crossover.Operator
	mutate    mutation.Operator
}

func (c *ChromosomeSpec) Crossover(op crossover.Operator) *ChromosomeSpec {
	if !op.IsCompatible(c.type_, uint(c.flags)) {
		// TODO either panic or issue a warning
		fmt.Println("WARN: incompatible crossover operator")
	}

	c.crossover = op
	return c
}

func (c *ChromosomeSpec) Mutate(op mutation.Operator) *ChromosomeSpec {
	if !op.IsCompatible(c.type_, uint(c.flags)) {
		// TODO either panic or issue a warning
		fmt.Println("WARN: incompatible mutation operator")
	}

	c.mutate = op
	return c
}

type GeneSpec struct {
	type_    reflect.Type
	isSlice  bool
	cells    int
	index    int
	len      int
	bits     int
	phOffset uintptr
}

var emptyGeneSpec = GeneSpec{}

func (g *GeneSpec) Index(i int) *GeneSpec {
	g.index = i
	return g
}
func (g *GeneSpec) Len(l int) *GeneSpec {
	g.len = l
	return g
}
func (g *GeneSpec) Bits(b int) *GeneSpec {
	if b > g.type_.Bits() {
		panic("specified bit width is too wide") // TODO
	}

	g.bits = b
	return g
}

type OpSpec struct{}

func Build[Phenotype any](spec func(bind BindFunc, ph *Phenotype) (s Spec)) (schema Schema[Phenotype], err error) {
	b := newBinder[Phenotype]()
	s := spec(b.bind, b.root)

	if err = b.validate(); err != nil {
		return
	}

	for _, cs := range s {
		c := Chromosome{
			type_:      cs.type_,
			bytesIndex: schema.sizeInBytes,
			crossover:  cs.crossover,
			mutate:     cs.mutate,
		}

		slices.SortStableFunc(cs.genes, func(a, b *GeneSpec) int {
			return b.bits - a.bits
		})

		var bf bestFitDecreasingAllocator

		for _, g := range cs.genes {
			bytes := g.type_.Align()

			var dynamic func(ptr, i uintptr) uintptr
			if g.isSlice {
				dynamic = func(ptr, i uintptr) uintptr {
					slice := *(*[]any)(unsafe.Pointer(ptr))
					dataptr := unsafe.Pointer(unsafe.SliceData(slice))
					return uintptr(unsafe.Add(dataptr, i))
				}

				for i := range g.len {
					index := uintptr((g.index + i) * bytes * g.cells)

					for j := range g.cells {
						locus := locus{
							bitWidth: g.bits,
						}
						bf.offer(&locus)

						gene := Gene{
							type_:           g.type_.Kind(), // TODO check compatibility and emit warnings
							phenotypeOffset: g.phOffset,
							dynamic:         dynamic,
							dynamicIndex:    index + uintptr(j*bytes),
							locus:           locus,
						}
						c.genes = append(c.genes, gene)
					}
				}
				continue
			}

			// TODO check len validity!
			for i := range g.len * g.cells {
				locus := locus{
					bitWidth: g.bits,
				}
				bf.offer(&locus)

				gene := Gene{
					type_:           g.type_.Kind(), // TODO check compatibility and emit warnings
					phenotypeOffset: g.phOffset + uintptr(i*bytes),
					locus:           locus,
				}

				c.genes = append(c.genes, gene)
			}
		}

		c.bytesLength = bf.nBytes()

		schema.chromosomes = append(schema.chromosomes, c)
		schema.sizeInBytes += c.bytesLength
	}
	return
}

func memEq[T any](a, b *T) bool {
	aBytes := unsafeBytes(a)
	bBytes := unsafeBytes(b)
	return bytes.Equal(aBytes, bBytes)
}
func unsafeBytes[T any](x *T) []byte {
	ptr := (*byte)(unsafe.Pointer(x))
	return unsafe.Slice(ptr, unsafe.Sizeof(*x))
}

type bestFitDecreasingAllocator struct {
	cellsFree []int
}

const cellSize = 64
const byteSize = 8
const bytesPerCell = cellSize / byteSize

func (bf bestFitDecreasingAllocator) nBytes() int {
	return len(bf.cellsFree)*bytesPerCell - bf.cellsFree[len(bf.cellsFree)-1]/byteSize
}
func (bf *bestFitDecreasingAllocator) offer(locus *locus) {

	bestFitIdx := -1
	minLeftover := cellSize + 1
	for i, space := range bf.cellsFree {
		left := space - locus.bitWidth
		if left >= 0 && (left < minLeftover) {
			bestFitIdx = i
			minLeftover = left
		}
	}

	if bestFitIdx > -1 {
		cellOffset := cellSize - bf.cellsFree[bestFitIdx]
		locus.byteIndex = bestFitIdx*bytesPerCell + cellOffset/byteSize
		locus.bitOffset = cellOffset % byteSize
		bf.cellsFree[bestFitIdx] -= locus.bitWidth
	} else {
		locus.byteIndex = len(bf.cellsFree) * bytesPerCell
		bf.cellsFree = append(bf.cellsFree, cellSize-locus.bitWidth)
	}
}
