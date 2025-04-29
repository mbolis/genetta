package mutation_test

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/mbolis/genetta/mutation"
	"github.com/stretchr/testify/assert"
)

const repeats = 10_000

func TestBitString(t *testing.T) {
	for _, n := range []int{1, 2, 4, 8, 16, 32} {
		t.Run(fmt.Sprintf("should flip %d bits per run on average", n), func(t *testing.T) {
			bs := mutation.BitString(n)

			var flips int
			for range repeats {
				genome := []byte{0, 0, 0, 0}
				err := bs.Mutate(genome)
				assert.NoError(t, err)

				for _, b := range genome {
					flips += bits.OnesCount8(b)
				}
			}

			epsilon := 0.01
			assert.InEpsilon(t, float64(flips)/float64(repeats), n, epsilon)
		})
	}
}
