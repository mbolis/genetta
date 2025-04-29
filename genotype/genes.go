package genotype

import (
	"fmt"
	"reflect"
	"unsafe"
)

type Gene struct {
	type_ reflect.Kind
	locus

	phenotypeOffset uintptr
	dynamic         dynamicPtr
	dynamicIndex    uintptr
}
type locus struct {
	byteIndex int
	bitOffset int
	bitWidth  int
}

type dynamicPtr func(ptr, i uintptr) uintptr

type integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64
}

func IntGene[T integer](bits int) (g Gene) {
	t := reflect.TypeFor[T]()
	if t.Bits() < bits {
		panic(fmt.Sprintf("type %v is too narrow: need %d bits", t, bits))
	}

	g.type_ = t.Kind()
	g.locus.bitWidth = bits
	return
}

var _ ChromosomeComponent = Gene{}

func (g Gene) apply(c *Chromosome) {
	c.genes = append(c.genes, g)
}

func (g Gene) Encode(ptr unsafe.Pointer, data []byte) {
	position := unsafe.Pointer(uintptr(ptr) + g.phenotypeOffset)
	if g.dynamic != nil {
		position = unsafe.Pointer(g.dynamic(uintptr(position), g.dynamicIndex))
	}

	var value uint64
	switch g.type_ {
	case reflect.Bool:
		read[bool](position, &value)
	case reflect.Int:
		read[int](position, &value)
	case reflect.Int8:
		read[int8](position, &value)
	case reflect.Int16:
		read[int16](position, &value)
	case reflect.Int32:
		read[int32](position, &value)
	case reflect.Int64:
		read[int64](position, &value)
	case reflect.Uint:
		read[uint](position, &value)
	case reflect.Uint8:
		read[uint8](position, &value)
	case reflect.Uint16:
		read[uint16](position, &value)
	case reflect.Uint32:
		read[uint32](position, &value)
	case reflect.Uint64:
		read[uint64](position, &value)
	case reflect.Float32:
		read[float32](position, &value)
	case reflect.Float64:
		read[float64](position, &value)
	}

	g.write(data, value)
}
func read[T any](position unsafe.Pointer, value *uint64) {
	*(*T)(unsafe.Pointer(value)) = *(*T)(position)
}

func (g Gene) Decode(ptr unsafe.Pointer, data []byte) {
	position := unsafe.Pointer(uintptr(ptr) + g.phenotypeOffset)
	if g.dynamic != nil {
		position = unsafe.Pointer(g.dynamic(uintptr(position), g.dynamicIndex))
	}

	value := g.read(data)

	switch g.type_ {
	case reflect.Bool:
		write[bool](position, value)
	case reflect.Int:
		write[int](position, value)
	case reflect.Int8:
		write[int8](position, value)
	case reflect.Int16:
		write[int16](position, value)
	case reflect.Int32:
		write[int32](position, value)
	case reflect.Int64:
		write[int64](position, value)
	case reflect.Uint:
		write[uint](position, value)
	case reflect.Uint8:
		write[uint8](position, value)
	case reflect.Uint16:
		write[uint16](position, value)
	case reflect.Uint32:
		write[uint32](position, value)
	case reflect.Uint64:
		write[uint64](position, value)
	case reflect.Float32:
		write[float32](position, value)
	case reflect.Float64:
		write[float64](position, value)
	}
}
func write[T any](position unsafe.Pointer, value uint64) {
	*(*T)(position) = *(*T)(unsafe.Pointer(&value))
}

const mask = ^uint64(0)

func (l locus) read(data []byte) (value uint64) {
	value = *(*uint64)(unsafe.Pointer(&data[l.byteIndex]))
	value >>= l.bitOffset
	value &= ^(mask << l.bitWidth)
	return
}

func (l locus) write(data []byte, value uint64) {
	mask := ^(mask << l.bitWidth)
	value &= mask

	target := (*uint64)(unsafe.Pointer(&data[l.byteIndex]))
	*target &^= mask << l.bitOffset
	*target |= value << l.bitOffset
}
