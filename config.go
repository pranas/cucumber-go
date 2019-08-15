package cucumber

// Configuration options
type Config struct {
	// Language (default "en")
	Language string

	// It is a good idea to randomize scenario order to catch
	// state dependency issues. (default cucumber.OrderRandom)
	Order OrderType

	// By default a random seed will be assigned,
	// assign any other value to reproduce scenario ordering
	// with particular seed
	Seed uint64

	// Max number of steps to run in parallel
	// 0 (default) - unbound
	// Use 1 to disable parallel execution
	Concurrency uint64

	// Stop on first failure
	FailFast bool

	// Do not execute steps
	DryRun bool

	// Fail on pending or undefined steps
	Strict bool

	// Filter scenarios by tags
	TagExpression string

	// By default it will use dot formatter configured to std out
	Formatter Formatter

	// By default it will look in features/ dir
	Paths []string
}
