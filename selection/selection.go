package selection

import (
	"fmt"
	"math/rand/v2"

	"github.com/mbolis/genetta/model"
)

type Operator interface {
	SelectInto(model.Genomes, [][]byte) error
}

type random struct{}

func Random() Operator {
	return random{}
}

func (random) SelectInto(p model.Genomes, buffer [][]byte) error {
	nIndividuals := p.NIndividuals()
	for i := range buffer {
		buffer[i] = p.Genotype(rand.IntN(nIndividuals))
	}
	return nil
}

type rouletteWheel struct{}

func RouletteWheel() Operator {
	return rouletteWheel{}
}

func (rouletteWheel) SelectInto(p model.Genomes, buffer [][]byte) error {
	p.MakeFitnessPositive()

	stats := p.Stats()
	totalFitness := stats.TotalFitness
	if totalFitness == 0 {
		return random{}.SelectInto(p, buffer)
	}

	nIndividuals := p.NIndividuals()
outer:
	for i := range buffer {
		selection := rand.Float64() * totalFitness
		initial := selection
		for idx := range nIndividuals {
			selection -= p.Fitness(idx)
			if selection <= 0 {
				buffer[i] = p.Genotype(idx)
				continue outer
			}
		}
		return fmt.Errorf("bad total fitness value: %f, selection: from %f down to %f", totalFitness, initial, selection)
	}
	return nil
}
