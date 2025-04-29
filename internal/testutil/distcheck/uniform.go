package distcheck

import (
	"fmt"
	"testing"

	"github.com/mbolis/genetta/internal/testutil/stats"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat/combin"
)

type Checker[T any] interface {
	Offer(T)
	Assert(*testing.T)
}

type uniformChecker struct {
	minValue int
	counts   []float64
	n        int
}

func Uniform(minValue, maxValue int) Checker[int] {
	return &uniformChecker{
		minValue: minValue,
		counts:   make([]float64, 1+maxValue-minValue),
	}
}

func (uc *uniformChecker) Offer(value int) {
	uc.counts[value-uc.minValue]++
	uc.n++
}

func (uc uniformChecker) Assert(t *testing.T) {
	t.Helper()

	nBins := float64(len(uc.counts))
	expected := float64(uc.n) / nBins
	_, pValue := stats.UniformChiSquareP(uc.counts, expected)
	assert.Greater(t, pValue, 0.05)
}

type nestedUniformChecker struct {
	size     int
	picks    int
	minValue int
	maxValue int
	counts   []float64
	n        float64

	tuplesByRank map[int][]int
}

func NestedUniform(picks, minValue, maxValue int) Checker[[]int] {
	return &nestedUniformChecker{
		size:         1 + maxValue - minValue,
		picks:        picks,
		minValue:     minValue,
		maxValue:     maxValue,
		counts:       make([]float64, combs(1+maxValue-minValue, picks)),
		tuplesByRank: make(map[int][]int),
	}
}

func (nuc *nestedUniformChecker) Offer(tuple []int) {
	if len(tuple) != nuc.picks {
		panic(fmt.Sprintf("bad tuple %d length must be %d", tuple, nuc.picks))
	}

	nuc.counts[nuc.rank(tuple)]++
	nuc.n++
}

func (nuc nestedUniformChecker) Assert(t *testing.T) {
	t.Helper()

	_, pValue := stats.ChiSquarePFunc(nuc.counts, func(i int) float64 {
		tuple := nuc.unrank(i)
		return nuc.n * nuc.prob(tuple)
	})
	assert.Greater(t, pValue, 0.05)
}

func (nuc nestedUniformChecker) rank(tuple []int) (rank int) {
	for i, v := range tuple {
		rank += combs(v-nuc.minValue, i+1)
	}
	nuc.tuplesByRank[rank] = tuple
	return
}

func (nuc nestedUniformChecker) unrank(rank int) (tuple []int) {
	return nuc.tuplesByRank[rank]
}

func (nuc nestedUniformChecker) prob(tuple []int) (p float64) {
	p = 1.0 / float64(nuc.size-nuc.picks+1)
	for i := 1; i < nuc.picks; i++ {
		p /= float64(nuc.size - tuple[i-1] - nuc.picks + 1 + i)
	}
	return
}

type combinations struct {
	minValue int
	n, k     int
	samples  float64
	eCount   []float64
	tCount   []float64
	tUnrank  [][]int
}

func Combinations(minValue, n, k int) Checker[[]int] {
	return &combinations{
		minValue: minValue,
		n:        n,
		k:        k,
		eCount:   make([]float64, n),
		tCount:   make([]float64, combs(n, k)),
		tUnrank:  make([][]int, combs(n, k)),
	}
}

func (c *combinations) Offer(t []int) {
	if len(t) != c.k {
		panic(fmt.Sprintf("bad tuple %d: length must be %d", t, c.k))
	}

	for _, e := range t {
		c.eCount[e-c.minValue]++
	}
	c.tCount[c.rank(t)]++
	c.samples++
}

func (c combinations) AssertElementDistribution(t *testing.T) {
	t.Helper()

	expected := c.samples * float64(c.k) / float64(c.n)
	_, pValue := stats.UniformChiSquareP(c.eCount, expected)
	assert.Greater(t, pValue, 0.05, "Elements should be equally distributed as T * k/n")
}

func (c combinations) AssertTupleDistribution(t *testing.T) {
	t.Helper()

	tot := float64(combs(c.n, c.k))
	expected := c.samples / tot
	_, pValue := stats.UniformChiSquareP(c.tCount, expected)
	assert.Greater(t, pValue, 0.05, "Elements should be equally distributed as T / C(n k)")
}

func (c *combinations) Assert(t *testing.T) {
	t.Helper()
	c.AssertElementDistribution(t)
	c.AssertTupleDistribution(t)
}

func (c combinations) rank(tuple []int) (rank int) {
	for i, v := range tuple {
		rank += combs(v-c.minValue, i+1)
	}
	c.tUnrank[rank] = tuple
	return
}

func combs(n, k int) int {
	if n < k {
		return 0
	}
	return combin.Binomial(n, k)
}
