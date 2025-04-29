package stats

import (
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

func ChiSquareP(observed []float64, expected []float64) (chi2, p float64) {
	chi2 = stat.ChiSquare(observed, expected)
	dist := distuv.ChiSquared{K: float64(len(observed) - 1)}
	return chi2, dist.CDF(chi2)
}

func ChiSquarePFunc(observed []float64, expectedFn func(int) float64) (chi2, p float64) {
	n := len(observed)
	if n == 0 {
		return 0, 1
	}

	for i, o := range observed {
		expected := expectedFn(i)
		diff := o - expected
		chi2 += diff * diff / expected
	}

	dist := distuv.ChiSquared{K: float64(len(observed) - 1)}
	return chi2, 1 - dist.CDF(chi2)
}

func UniformChiSquareP(observed []float64, expected float64) (chi2, p float64) {
	n := len(observed)
	if n == 0 {
		return 0, 1
	}

	for _, o := range observed {
		diff := o - expected
		chi2 += diff * diff / expected
	}

	dist := distuv.ChiSquared{K: float64(len(observed) - 1)}
	return chi2, 1 - dist.CDF(chi2)
}
