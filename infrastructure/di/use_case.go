package di

import "github.com/Pimousse1099/fizz-buzz-api/usecase"

// getGenerateFizzBuzzUseCase returns the memoized generate use-case.
func (c *Container) getGenerateFizzBuzzUseCase() *usecase.GenerateFizzBuzz {
	if c.generateFizzBuzzUseCase == nil {
		c.generateFizzBuzzUseCase = usecase.NewGenerateFizzBuzz(
			c.config.FizzBuzz.MaxSequenceLength,
			c.getStatStore(),
		)
	}

	return c.generateFizzBuzzUseCase
}

// getFizzBuzzStatsUseCase returns the memoized stats use-case.
func (c *Container) getFizzBuzzStatsUseCase() *usecase.GetFizzBuzzStats {
	if c.fizzBuzzStatsUseCase == nil {
		c.fizzBuzzStatsUseCase = usecase.NewGetFizzBuzzStats(c.getStatStore())
	}

	return c.fizzBuzzStatsUseCase
}
