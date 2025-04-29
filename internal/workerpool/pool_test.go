package workerpool_test

import (
	"slices"
	"testing"

	"github.com/mbolis/genetta/internal/util"
	"github.com/mbolis/genetta/internal/workerpool"
	"github.com/stretchr/testify/assert"
)

type testWorker struct {
	count int
	sum   int
}

func (t *testWorker) work(i int) {
	t.count++
	t.sum += i
}

const (
	nWorkers = 8
	jobs     = 1_000_000
)

func TestPool(t *testing.T) {
	t.Run("should run tasks in parallel", func(t *testing.T) {
		pool, err := workerpool.New(nWorkers, jobs, (*testWorker).work)
		assert.NoError(t, err)
		defer pool.Close()

		var expectedSum int
		for i := range jobs {
			pool.Offer(i)
			expectedSum += i
		}
		pool.Wait()

		var counts [nWorkers]float64
		var actualSum int
		for i, w := range pool.All() {
			counts[i] = float64(w.count)
			actualSum += w.sum
		}

		assert.Equal(t, expectedSum, actualSum)
		assert.InDeltaSlice(t,
			slices.Repeat([]float64{1.0 / nWorkers}, nWorkers),
			util.Map(counts[:], func(c float64) float64 { return c / jobs }),
			0.005,
		)
	})

	t.Run("should run tasks in two batches with resume", func(t *testing.T) {
		pool, err := workerpool.New(int(nWorkers), jobs, (*testWorker).work)
		assert.NoError(t, err)
		defer pool.Close()

		var expectedSum int
		for i := range jobs {
			pool.Offer(i)
			expectedSum += i
		}
		pool.Wait()

		var actualSum int
		for _, w := range pool.All() {
			actualSum += w.sum
		}
		assert.Equal(t, expectedSum, actualSum)

		for i := range jobs {
			pool.Offer(i)
		}
		pool.Resume()
		pool.Wait()

		actualSum = 0
		for _, w := range pool.All() {
			actualSum += w.sum
		}
		assert.Equal(t, expectedSum, actualSum)
	})
}
