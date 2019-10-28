# Cucumber for Go

Cucumber runner for Go with focus on concurrency. It supports and encourages isolating scenario state and running scenarios in parallel.

The package is based on the official [cucumber-engine](https://github.com/cucumber/cucumber-engine) from Cucumber team.

## Configuration

The following configuration options are available.

```golang
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
```

## Usage

You would typically create `cmd/cucumber/cucumber.go` similar to this:

```golang
func main() {
    s, err := cucumber.NewSuite(cucumber.Config{}, os.Args[1:]...)
    if err != nil {
        panic(err)
    }

    s.DefineTestCaseInitializer(func(tc cucumber.TestCase) {
        tc.Set("state", "foo")
    })

    s.DefineStep(`^state should be "([^"]*)"$`, func(tc cucumber.TestCase, matches ...string) error {
        currentState := tc.Get("state").(string)
        if currentState != matches[0] {
            return fmt.Errorf("expected %s but got %s", matches[0], currentState)
        }
        return nil
    })

    s.DefineStep(`^you should have "([^"]*)"$`, func(tc cucumber.TestCase, matches ...string) error {
        return cucumber.ErrPending
    })

    exitCode := s.Run()
    os.Exit(exitCode)
}
```

## TODO

* Pretty formatter
* Support Cucumber expressions (currently regex)
* godoc

## License

This repository is licensed under the MIT license.
