package genetta

import (
	"fmt"
	"math"

	"github.com/mbolis/genetta/genotype"
	"github.com/mbolis/genetta/model"
	"github.com/mbolis/genetta/selection"
)

type GA[Phenotype any] interface {
	Epoch() (Result[Phenotype], bool)
	Epochs(int) (Result[Phenotype], bool)

	Generation() int
	TargetFitness() float64
	Elitism() (size, copies int)
}

type gaSolver[P any] struct {
	schema       genotype.Schema[P]
	population   model.Population[P]
	breedingPool [][]byte
	generation   int

	fitnessFunc func(P) float64
	opts        options
}

type options struct {
	populationSize int
	targetFitness  *float64
	selectionOp    selection.Operator
	// scalingOp scaling.Operator
	elite struct {
		size   int
		copies int
		len    int
	}
	// constraints []constraint.Constraint[P] // XXX in schema...
}

type Option func(*options) error

func WithTargetFitness(f float64) func(*options) error {
	return func(o *options) error {
		o.targetFitness = &f
		return nil
	}
}
func WithSelection(op selection.Operator) func(*options) error {
	return func(o *options) error {
		o.selectionOp = op
		return nil
	}
}
func WithElitism(size, copies int) func(*options) error {
	return func(o *options) error {
		if size <= 0 || copies <= 0 {
			return fmt.Errorf("elite size/copies must be > 0, was %d/%d", size, copies)
		}

		o.elite.size = size
		o.elite.copies = copies
		o.elite.len = size * copies
		return nil
	}
}

func NewSolver[P any](genotype genotype.Schema[P], fitnessFunc func(P) float64, populationSize int, opts ...Option) (GA[P], error) {
	if populationSize <= 0 {
		return nil, fmt.Errorf("population size must be > 0, was %d", populationSize)
	}

	o := options{
		populationSize: populationSize,
	}
	for _, opt := range opts {
		err := opt(&o)
		if err != nil {
			return nil, err
		}
	}

	return &gaSolver[P]{
		schema:       genotype,
		fitnessFunc:  fitnessFunc,
		opts:         o,
		population:   model.New(genotype, populationSize),
		breedingPool: make([][]byte, populationSize),
		generation:   1,
	}, nil
}

func (ga gaSolver[P]) Generation() int {
	return ga.generation
}
func (ga gaSolver[P]) TargetFitness() float64 {
	if ga.opts.targetFitness == nil {
		return math.NaN()
	}
	return *ga.opts.targetFitness
}
func (ga gaSolver[P]) Elitism() (size, copy int) {
	return ga.opts.elite.size, ga.opts.elite.copies
}

type Result[P any] struct {
	population model.Population[P]
	index      int
}

func (ga *gaSolver[P]) Epoch() (fittest Result[P], found bool) {
	return ga.Epochs(1)
}

func (ga *gaSolver[P]) Epochs(n int) (fittest Result[P], found bool) {
	for range n {
		fittest, found = ga.calculateFitnessScores()
		if found {
			break
		}
		ga.nextGeneration()
	}
	return
}

func (ga *gaSolver[P]) calculateFitnessScores() (Result[P], bool) {
	ga.population.Reset()

	phenotype := ga.schema.Init()
	for i := range ga.population.NIndividuals() { // TODO parallelize
		ga.population.Decode(i, &phenotype)
		ga.population.SetFitness(i, ga.calculateFitness(phenotype))
	}

	fittest, maxFitness := ga.population.Fittest()
	result := Result[P]{ga.population, fittest}
	if ga.opts.targetFitness != nil && maxFitness == *ga.opts.targetFitness {
		return result, true
	}
	// TODO scale fitness
	return result, false
}

func (ga *gaSolver[P]) calculateFitness(phenotype P) float64 {
	return ga.fitnessFunc(phenotype)
}

func (ga *gaSolver[P]) nextGeneration() {
	// TODO elite

	ga.selectBreedingPool()
	for i := 0; i < ga.population.NIndividuals(); i += 2 { // TODO parallelize
		mom := ga.breedingPool[i]
		dad := ga.breedingPool[i+1]

		child1 := ga.population.Genotype(i)
		child2 := ga.population.Genotype(i + 1)

		if err := ga.schema.Crossover(mom, dad, child1, child2); err != nil {
			// TODO
		}
		if err := ga.schema.Mutate(child1, child2); err != nil {
			// TODO
		}
		// TODO enforce constraints...
	}

	ga.generation++
}

func (ga *gaSolver[P]) selectBreedingPool() {
	var pos int
	if ga.opts.elite.len > 0 {
		ga.population.SortByFitnessDesc()

		for i := range ga.opts.elite.size {
			for range ga.opts.elite.copies {
				ga.breedingPool[pos] = ga.population.Genotype(i)
				pos++
			}
		}
	}

	if err := ga.opts.selectionOp.SelectInto(ga.population.Genomes, ga.breedingPool[pos:]); err != nil {
		// TODO
	}
}

// type Population[P any] struct {
// 	schema genotype.Schema[P]
// 	size   int

// 	data        []byte
// 	individuals []Individual[P]

// 	selectionIndividuals []selection.Individual
// 	isSorted             bool

// 	fittest      *Individual[P]
// 	worst        *Individual[P]
// 	totalFitness float64
// 	variance     float64

// 	generation int
// }

// var _ selection.Population = (*Population[any])(nil)

// func newPopulation[P any](schema genotype.Schema[P], size int) Population[P] {
// 	p := Population[P]{
// 		schema:               schema,
// 		size:                 size,
// 		data:                 schema.Make(size),
// 		individuals:          make([]Individual[P], size),
// 		selectionIndividuals: make([]selection.Individual, size),
// 		generation:           1,
// 	}

// 	for i := range p.individuals {
// 		idx1, idx2 := schema.Bounds(i)

// 		p.individuals[i] = Individual[P]{
// 			schema:   &p.schema,
// 			genotype: p.data[idx1:idx2],
// 		}

// 		p.individuals[i].Randomize()

// 		p.selectionIndividuals[i] = &p.individuals[i]
// 	}

// 	return p
// }

// func (p Population[P]) Generation() int {
// 	return p.generation
// }
// func (p Population[P]) Individuals() []selection.Individual {
// 	return p.selectionIndividuals
// }
// func (p *Population[P]) MakeFitnessPositive() {
// 	worstFitness := p.worst.fitness
// 	if worstFitness >= 0 {
// 		return
// 	}

// 	for i := range p.individuals {
// 		p.individuals[i].fitness -= worstFitness
// 	}
// 	p.totalFitness -= worstFitness * float64(p.size)
// }
// func (i *Population[P]) SortByFitnessDesc() {
// 	if i.isSorted {
// 		return
// 	}

// 	slices.SortStableFunc(i.individuals, func(a, b Individual[P]) int {
// 		return cmp.Compare(b.fitness, a.fitness)
// 	})
// 	i.isSorted = true
// }

// func (p Population[P]) Stats() selection.Stats {
// 	return p
// }
// func (p Population[P]) Fittest() selection.Individual {
// 	return p.fittest
// }
// func (p Population[P]) Worst() selection.Individual {
// 	return p.worst
// }
// func (p Population[P]) TotalFitness() float64 {
// 	return p.totalFitness
// }
// func (p Population[P]) Mean() float64 {
// 	return p.totalFitness / float64(p.size)
// }
// func (p Population[P]) Variance() float64 {
// 	if math.IsNaN(p.variance) {
// 		mean := p.Mean()
// 		p.variance = 0
// 		for _, i := range p.individuals {
// 			delta := i.fitness - mean
// 			p.variance += delta * delta
// 		}
// 		p.variance /= float64(p.size)
// 	}

// 	return p.variance
// }
// func (p Population[P]) StandardDeviation() float64 {
// 	return math.Sqrt(p.Variance())
// }

// func (p Population[P]) All() iter.Seq[*Individual[P]] {
// 	return p.iterator
// }
// func (p Population[P]) iterator(yield func(*Individual[P]) bool) {
// 	for i := range p.individuals {
// 		if !yield(&p.individuals[i]) {
// 			break
// 		}
// 	}
// }

// type Individual[P any] struct {
// 	schema   *genotype.Schema[P]
// 	genotype []byte
// 	fitness  float64
// }

// func (i Individual[P]) Genotype() []byte {
// 	return i.genotype
// }
// func (i *Individual[P]) Fitness() float64 {
// 	if i == nil {
// 		return 0
// 	}
// 	return i.fitness
// }

// func (i Individual[P]) Decode() (phenotype P) {
// 	i.schema.Decode(&phenotype, i.genotype)
// 	return
// }

// func (i Individual[P]) Randomize() {
// 	i.schema.Randomize(i.genotype)
// }
