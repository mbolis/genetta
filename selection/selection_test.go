package selection_test

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/mbolis/genetta/genotype"
	"github.com/mbolis/genetta/internal/testutil/distcheck"
	"github.com/mbolis/genetta/internal/testutil/stats"
	"github.com/mbolis/genetta/internal/util"
	"github.com/mbolis/genetta/model"
	"github.com/mbolis/genetta/selection"
	"github.com/stretchr/testify/assert"
)

const repeats = 10_000

func randomPopulation(size int) model.Population[[]byte] {
	p := model.New(genotype.Binary[byte](8, 1), size)
	for i := range p.NIndividuals() {
		p.Encode(i, []byte{byte(i + 1)})
		p.SetFitness(i, math.Abs(rand.NormFloat64()))
	}
	return p
}

func TestRandom(t *testing.T) {
	r := selection.Random()

	uniform := distcheck.Uniform(1, 128)
	for range repeats {
		var buffer [128][]byte

		population := randomPopulation(128)
		err := r.SelectInto(population.Genomes, buffer[:])
		assert.NoError(t, err)

		for _, g := range buffer {
			uniform.Offer(int(g[0]))
		}
	}

	uniform.Assert(t)
}

func TestRouletteWheel(t *testing.T) {
	r := selection.RouletteWheel()

	var counts [128]float64
	population := randomPopulation(128)
	for range repeats {
		var buffer [128][]byte

		err := r.SelectInto(population.Genomes, buffer[:])
		assert.NoError(t, err)

		for _, g := range buffer {
			counts[g[0]-1]++
		}
	}

	tot := population.Stats().TotalFitness * 128 * repeats
	_, pValue := stats.ChiSquareP(
		counts[:],
		util.Range(population.NIndividuals(), func(i int) float64 {
			return population.Fitness(i) / tot
		}),
	)
	assert.Greater(t, pValue, 0.05)
}
