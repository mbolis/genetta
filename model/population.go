package model

import (
	"math"
	"sort"

	"github.com/mbolis/genetta/genotype"
)

type Genomes struct {
	size          int
	chromosomeLen int
	genotype      []byte

	fitness      []float64
	totalFitness float64

	isSorted bool

	fittest int
	worst   int
}

type GenomesSortDesc struct {
	*Genomes
}

func (g GenomesSortDesc) Len() int {
	return g.size
}
func (g GenomesSortDesc) Less(i, j int) bool {
	return g.fitness[i] > g.fitness[j]
}
func (g GenomesSortDesc) Swap(i, j int) {
	g.fitness[i], g.fitness[j] = g.fitness[j], g.fitness[i]

	i0 := i * g.chromosomeLen
	j0 := j * g.chromosomeLen
	for x := range g.chromosomeLen {
		i := i0 + x
		j := j0 + x
		g.genotype[i], g.genotype[j] = g.genotype[j], g.genotype[i]
	}
}

func (g Genomes) NIndividuals() int {
	return g.size
}

func (g Genomes) Genotype(i int) []byte {
	o := i * g.chromosomeLen
	return g.genotype[o : o+g.chromosomeLen]
}

func (g Genomes) Fitness(i int) float64 {
	return g.fitness[i]
}
func (g *Genomes) SetFitness(i int, f float64) { // XXX better to produce aggregates in the parallel code, then set
	g.totalFitness += f - g.fitness[i]
	g.fitness[i] = f

	if g.worst < 0 || f < g.fitness[g.worst] {
		g.worst = i
	}
	if g.fittest < 0 || f > g.fitness[g.fittest] {
		g.fittest = i
	}
}

func (g Genomes) Fittest() (int, float64) {
	return g.fittest, g.fitness[g.fittest]
}
func (g Genomes) Worst() (int, float64) {
	return g.worst, g.fitness[g.worst]
}

func (g *Genomes) MakeFitnessPositive() {
	worstFitness := g.fitness[g.worst]
	if worstFitness >= 0 {
		return
	}

	for i := range g.size {
		g.fitness[i] -= worstFitness
	}
	g.totalFitness -= worstFitness * float64(g.size)
}

func (g *Genomes) SortByFitnessDesc() {
	if g.isSorted {
		return
	}

	sort.Stable(GenomesSortDesc{g})
	g.isSorted = true
}

type Stats struct {
	Fittest int
	Worst   int

	MinFitness float64
	MaxFitness float64

	NValues      float64
	TotalFitness float64
	Mean         float64

	fitnessValues []float64
}

func (p Genomes) Stats() Stats {
	return Stats{
		Fittest:       p.fittest,
		Worst:         p.worst,
		MinFitness:    p.fitness[p.worst],
		MaxFitness:    p.fitness[p.fittest],
		NValues:       float64(p.size),
		TotalFitness:  p.totalFitness,
		Mean:          p.totalFitness / float64(p.size),
		fitnessValues: p.fitness,
	}
}
func (s Stats) Variance() float64 {
	mean := s.Mean
	variance := 0.0
	for _, f := range s.fitnessValues {
		delta := f - mean
		variance += delta * delta
	}
	return variance / s.NValues
}
func (s Stats) StandardDeviation() float64 {
	return math.Sqrt(s.Variance())
}

type Population[P any] struct {
	schema genotype.Schema[P]
	Genomes
}

func New[P any](schema genotype.Schema[P], size int) Population[P] {
	p := Population[P]{
		schema,
		Genomes{
			size:          size,
			genotype:      schema.Make(size),
			fitness:       make([]float64, size),
			chromosomeLen: schema.Size(),
		},
	}
	p.Randomize()
	return p
}

func (p Population[P]) Randomize() {
	for i := range p.size {
		p.schema.Randomize(p.Genotype(i))
	}
}

func (p Population[P]) Encode(i int, phenotype P) {
	p.schema.Encode(&phenotype, p.Genotype(i))
}
func (p Population[P]) Decode(i int, phenotype *P) {
	p.schema.Decode(phenotype, p.Genotype(i))
}

func (p *Population[P]) Reset() {
	p.totalFitness = 0
	p.fittest = -1
	p.worst = -1
	p.isSorted = false
}
