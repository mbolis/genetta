package genotype_test

import (
	"math"
	"testing"

	"github.com/mbolis/genetta/genotype"
	"github.com/stretchr/testify/assert"
)

func TestBuildOptions(t *testing.T) {
	s, err := genotype.Build(func(bind genotype.BindFunc, ph *[10]int) (s genotype.Spec) {
		s.IntChromosome(
			bind(ph).Len(5).Bits(12),
		)
		s.IntChromosome(
			bind(&ph[5]).Len(5).Bits(4),
		)
		return
	})

	assert.NoError(t, err)

	genotype := []byte{0x11, 0x21, 0x22, 0x33, 0x43, 0x44, 0x55, 0x05, 0x21, 0x43, 0x5}
	var phenotype [10]int
	s.Decode(&phenotype, genotype)
	assert.Equal(t,
		[10]int{0x111, 0x222, 0x333, 0x444, 0x555, 0x1, 0x2, 0x3, 0x4, 0x5},
		phenotype,
	)
}

type TestStruct struct {
	b    bool
	ints struct {
		i   int
		i64 int64
		i32 int32
		i16 int16
		i8  int8
	}
	uints struct {
		u   uint
		u64 uint64
		u32 uint32
		u16 uint16
		u8  uint8
	}
	floats struct {
		f32  float32
		f64  float64
		c64  complex64
		c128 complex128
	}
	arrays struct {
		b [3]bool
		i [3]int16
		f [3]float32
		c [3]complex64
	}
}

func buildSchema(t assert.TestingT) genotype.Schema[TestStruct] {
	s, err := genotype.Build(func(bind genotype.BindFunc, ph *TestStruct) (s genotype.Spec) {
		s.IntChromosome(
			bind(&ph.b),
			bind(&ph.ints.i),
			bind(&ph.ints.i64),
			bind(&ph.ints.i32),
			bind(&ph.ints.i16),
			bind(&ph.ints.i8),
			bind(&ph.uints.u),
			bind(&ph.uints.u64),
			bind(&ph.uints.u32),
			bind(&ph.uints.u16),
			bind(&ph.uints.u8),
		)
		s.Float32Chromosome(
			bind(&ph.floats.f32),
			bind(&ph.floats.c64),
		)
		s.Float64Chromosome(
			bind(&ph.floats.f64),
			bind(&ph.floats.c128),
		)
		s.IntChromosome(
			bind(&ph.arrays.b),
			bind(&ph.arrays.i),
		)
		s.Float32Chromosome(
			bind(&ph.arrays.f),
			bind(&ph.arrays.c),
		)
		return
	})

	assert.NoError(t, err)

	return s
}

func TestBuildSchema(t *testing.T) {
	s := buildSchema(t)
	genotype := s.Make(1)

	var v TestStruct
	v.b = true
	v.ints.i = ^int(0)
	v.ints.i64 = ^int64(0)
	v.ints.i32 = ^int32(0)
	v.ints.i16 = ^int16(0)
	v.ints.i8 = ^int8(0)
	v.uints.u = ^uint(0)
	v.uints.u64 = ^uint64(0)
	v.uints.u32 = ^uint32(0)
	v.uints.u16 = ^uint16(0)
	v.uints.u8 = ^uint8(0)
	v.floats.f32 = float32(math.NaN())
	v.floats.f64 = math.NaN()
	v.floats.c64 = complex(v.floats.f32, v.floats.f32)
	v.floats.c128 = complex(v.floats.f64, v.floats.f64)
	v.arrays.b = [3]bool{v.b, v.b, v.b}
	v.arrays.i = [3]int16{v.ints.i16, v.ints.i16, v.ints.i16}
	v.arrays.f = [3]float32{v.floats.f32, v.floats.f32, v.floats.f32}
	v.arrays.c = [3]complex64{v.floats.c64, v.floats.c64, v.floats.c64}

	s.Encode(&v, genotype)

	assert.Equal(t, []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // int
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // int64
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // uint
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // uint64
		0xff, 0xff, 0xff, 0xff, // int32
		0xff, 0xff, 0xff, 0xff, // uint32
		0xff, 0xff, // int16
		0xff, 0xff, // uint16
		0xff, // int8
		0xff, // uint8
		0x01, // bool

		0x00, 0x00, 0xc0, 0x7f, // float32
		0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, // complex64

		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f, // float64
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f, // complex128

		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // [3]int16
		0x07, // [3]bool

		0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, // [3]float32
		0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, 0x00, 0x00, 0xc0, 0x7f, // [3]complex64
	}, genotype)
}

func FuzzBuildInts(f *testing.F) {
	f.Add(true, 123, int64(-8765432), int32(3456789), int16(2345), int8(-123))

	s := buildSchema(f)
	genotype := s.Make(1)

	f.Fuzz(func(t *testing.T, b bool, i int, i64 int64, i32 int32, i16 int16, i8 int8) {
		var v TestStruct
		v.b = b
		v.ints.i = i
		v.ints.i64 = i64
		v.ints.i32 = i32
		v.ints.i16 = i16
		v.ints.i8 = i8

		s.Encode(&v, genotype)

		var d TestStruct
		s.Decode(&d, genotype)
		assert.Equal(t, v, d)
	})
}

func FuzzBuildUints(f *testing.F) {
	f.Add(uint(4567), uint64(987654321), uint32(4567890), uint16(34567), uint8(234))

	s := buildSchema(f)
	genotype := s.Make(1)

	f.Fuzz(func(t *testing.T, u uint, u64 uint64, u32 uint32, u16 uint16, u8 uint8) {
		var v TestStruct
		v.uints.u = u
		v.uints.u64 = u64
		v.uints.u32 = u32
		v.uints.u16 = u16
		v.uints.u8 = u8

		s.Encode(&v, genotype)

		var d TestStruct
		s.Decode(&d, genotype)
		assert.Equal(t, v, d)
	})
}

func FuzzBuildFloats(f *testing.F) {
	f.Add(float32(0.123), float64(-4.567), float32(-8.9), float32(0.1), 23.4, -5.67)

	s := buildSchema(f)
	genotype := s.Make(1)

	f.Fuzz(func(t *testing.T, f32 float32, f64 float64, c64r, c64i float32, c128r, c128i float64) {
		var v TestStruct
		v.floats.f32 = f32
		v.floats.f64 = f64
		v.floats.c64 = complex(c64r, c64i)
		v.floats.c128 = complex(c128r, c128i)

		s.Encode(&v, genotype)

		var d TestStruct
		s.Decode(&d, genotype)
		assert.Equal(t, v, d)
	})
}

func FuzzBuildArrays(f *testing.F) {
	f.Add(
		true, false, true,
		int16(1), int16(2), int16(3),
		float32(4.5), float32(6.7), float32(8.9),
		float32(1.2), float32(3.4), float32(5.6), float32(7.8), float32(9), float32(0.1),
	)

	s := buildSchema(f)
	genotype := s.Make(1)

	f.Fuzz(func(t *testing.T, b1, b2, b3 bool, i1, i2, i3 int16, f1, f2, f3 float32, c1r, c1i, c2r, c2i, c3r, c3i float32) {
		var v TestStruct
		v.arrays.b = [3]bool{b1, b2, b3}
		v.arrays.i = [3]int16{i1, i2, i3}
		v.arrays.f = [3]float32{f1, f2, f3}
		v.arrays.c = [3]complex64{complex(c1r, c1i), complex(c2r, c2i), complex(c3r, c3i)}

		s.Encode(&v, genotype)

		var d TestStruct
		s.Decode(&d, genotype)
		assert.Equal(t, v, d)
	})
}

func TestBuildSlice(t *testing.T) {
	t.Run("should encode/decode slice", func(t *testing.T) {
		s, err := genotype.Build(func(bind genotype.BindFunc, ph *[]int) (s genotype.Spec) {
			s.IntChromosome(
				bind(ph).Len(3).Bits(2),
			)
			return
		})
		assert.NoError(t, err)

		v := []int{1, 2, 3, 4, 5}
		genotype := s.Make(1)
		s.Encode(&v, genotype)

		d := make([]int, 5)
		s.Decode(&d, genotype)
		assert.Equal(t, []int{1, 2, 3, 0, 0}, d)
	})
	t.Run("should encode/decode slice with index", func(t *testing.T) {
		s, err := genotype.Build(func(bind genotype.BindFunc, ph *[]int) (s genotype.Spec) {
			s.IntChromosome(
				bind(ph).Index(1).Len(3).Bits(3),
			)
			return
		})
		assert.NoError(t, err)

		v := []int{1, 2, 3, 4, 5}
		genotype := s.Make(1)
		s.Encode(&v, genotype)

		d := make([]int, 5)
		s.Decode(&d, genotype)
		assert.Equal(t, []int{0, 2, 3, 4, 0}, d)
	})
}
