package di

import "github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"

// getStatStore returns the memoized in-memory stat store (it satisfies both the
// recorder and reader interfaces consumed by the use-cases).
func (c *Container) getStatStore() *statstorer.InMemory {
	if c.statStore == nil {
		c.statStore = statstorer.NewInMemory()
	}

	return c.statStore
}
