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

// getGetFizzBuzzStatsUseCase returns the memoized stats use-case.
func (c *Container) getGetFizzBuzzStatsUseCase() *usecase.GetFizzBuzzStats {
	if c.getFizzBuzzStatsUseCase == nil {
		c.getFizzBuzzStatsUseCase = usecase.NewGetFizzBuzzStats(c.getStatStore())
	}

	return c.getFizzBuzzStatsUseCase
}
