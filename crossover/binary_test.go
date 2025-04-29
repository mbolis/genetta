package crossover_test

import (
	"fmt"
	"math/bits"
	"math/rand/v2"
	"testing"
	"unsafe"

	"github.com/mbolis/genetta/crossover"
	"github.com/mbolis/genetta/internal/testutil/distcheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const repeats = 10_000

func TestSinglePoint(t *testing.T) {
	sp := crossover.SinglePoint()
	mom := []byte{0xaa, 0xaa, 0xaa, 0xaa}
	dad := []byte{0x55, 0x55, 0x55, 0x55}

	minXP := 31
	maxXP := 0
	uniform := distcheck.Uniform(1, 30)

	for range repeats {
		var child1, child2 [4]byte

		err := sp.Crossover(mom, dad, child1[:], child2[:])
		assert.NoError(t, err)

		xps1 := findCrossoverPoints(0xaa, 0x55, child1[:])
		xps2 := findCrossoverPoints(0x55, 0xaa, child2[:])
		require.Len(t, xps1, 1)
		assert.Equal(t, xps1, xps2)

		xp := xps1[0]
		minXP = min(minXP, xp)
		maxXP = max(maxXP, xp)
		uniform.Offer(xp)
	}

	assert.Equal(t, 1, minXP)
	assert.Equal(t, 30, maxXP)
	uniform.Assert(t)
}

func TestTwoPoints(t *testing.T) {
	tp := crossover.TwoPoints()
	mom := []byte{0xaa, 0xaa, 0xaa, 0xaa}
	dad := []byte{0x55, 0x55, 0x55, 0x55}

	minXP := 31
	maxXP := 0
	uniform := distcheck.Combinations(1, 30, 2)

	for range repeats {
		var child1, child2 [4]byte

		err := tp.Crossover(mom, dad, child1[:], child2[:])
		assert.NoError(t, err)

		xps1 := findCrossoverPoints(0xaa, 0x55, child1[:])
		xps2 := findCrossoverPoints(0x55, 0xaa, child2[:])
		require.Len(t, xps1, 2)
		assert.Equal(t, xps1, xps2)
		assert.Less(t, xps1[0], xps1[1])

		minXP = min(minXP, xps1[0])
		maxXP = max(maxXP, xps1[1])
		uniform.Offer(xps1)
	}

	assert.Equal(t, 1, minXP)
	assert.Equal(t, 30, maxXP)
	uniform.Assert(t)
}

func TestKPoints(t *testing.T) {
	const k = 5

	kp := crossover.KPoints(k)
	mom := []byte{0xaa, 0xaa, 0xaa, 0xaa}
	dad := []byte{0x55, 0x55, 0x55, 0x55}

	minXP := 31
	maxXP := 0
	uniform := distcheck.Combinations(1, 30, k)

	for range repeats {
		var child1, child2 [4]byte

		err := kp.Crossover(mom, dad, child1[:], child2[:])
		assert.NoError(t, err)

		xps1 := findCrossoverPoints(0xaa, 0x55, child1[:])
		xps2 := findCrossoverPoints(0x55, 0xaa, child2[:])
		require.Len(t, xps1, 5)
		assert.Equal(t, xps1, xps2)
		assertStrongSorting(t, xps1)

		minXP = min(minXP, xps1[0])
		maxXP = max(maxXP, xps1[k-1])
		uniform.Offer(xps1)
	}

	assert.Equal(t, 1, minXP)
	assert.Equal(t, 30, maxXP)
	uniform.Assert(t)
}

func findCrossoverPoints(orig, cross byte, d []byte) (xps []int) {
	var xo bool
	expected := orig

	for i, b := range d {
		if b == expected {
			continue
		}

		for j := range 8 {
			mask := byte(1) << j
			if b&mask != expected&mask {
				xps = append(xps, i*8+j)

				xo = !xo
				if xo {
					expected = cross
				} else {
					expected = orig
				}
			}
		}
	}
	return
}

func assertStrongSorting[T any](t *testing.T, slice []T) {
	t.Helper()

	for i := 1; i < len(slice); i++ {
		assert.Less(t, slice[i-1], slice[i], "Slice elements should be strongly sorted")
	}
}

func TestParametricHalfUniform(t *testing.T) {
	for _, rate := range []float64{0.25, 0.5, 0.75} {
		t.Run(fmt.Sprintf("should flip %d%% of the times", int(rate*100)), func(t *testing.T) {
			pu := crossover.ParametricHalfUniform(rate)
			mom := []byte{0xaa, 0xaa, 0xaa, 0xaa}
			dad := []byte{0x55, 0x55, 0x55, 0x55}

			var flipped, equal int
			for range repeats {
				var child1, child2 [4]byte

				err := pu.Crossover(mom, dad, child1[:], child2[:])
				assert.NoError(t, err)

				for i := range child1 {
					equal += bits.OnesCount8(child1[i] & child2[i])
					flipped += bits.OnesCount8(child1[i] ^ mom[i])
					assert.Equal(t, child1[i]^mom[i], child2[i]^dad[i])
				}
			}

			assert.Zero(t, equal)

			totBits := repeats * len(mom) * 8
			epsilon := 0.01
			assert.InEpsilon(t, float64(flipped)/float64(totBits), rate, epsilon)
		})
	}
	for _, rate := range []float64{0.25, 0.5, 0.75} {
		t.Run(fmt.Sprintf("should flip %d%% of the times when not equal", int(rate*100)), func(t *testing.T) {
			pu := crossover.ParametricHalfUniform(rate)

			var flipped, equal int
			for range repeats {
				var mom, dad [4]byte
				var child1, child2 [4]byte

				*(*uint32)(unsafe.Pointer(&mom)) = rand.Uint32()
				*(*uint32)(unsafe.Pointer(&dad)) = rand.Uint32()

				err := pu.Crossover(mom[:], dad[:], child1[:], child2[:])
				assert.NoError(t, err)

				for i := range child1 {
					equal += 8 - bits.OnesCount8(child1[i]^child2[i])
					assert.Equal(t, child1[i]^child2[i], mom[i]^dad[i])

					flipped += bits.OnesCount8(child1[i] ^ mom[i])
					assert.Equal(t, child1[i]^mom[i], child2[i]^dad[i])
				}
			}

			totBits := repeats * 32
			epsilon := 0.01
			assert.InEpsilon(t, float64(equal)/float64(totBits), 0.5, epsilon)
			assert.InEpsilon(t, float64(flipped)/float64(totBits), rate*0.5, epsilon)
		})
	}
}
