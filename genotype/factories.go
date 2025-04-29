package genotype

func Binary[T integer](bits, len int) Schema[[]T] {
	s, err := Build[[]T](func(bind BindFunc, ph *[]T) (s Spec) {
		s.IntChromosome(bind(ph).Bits(bits).Len(len))
		return
	})
	if err != nil {
		panic(err)
	}
	return s
}
