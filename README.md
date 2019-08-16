# Cucumber for Go

Cucumber runner for Go with focus on concurrency. It supports and encourages isolating scenario state and running scenarios in parallel.

The package is based on the official [cucumber-engine](https://github.com/cucumber/cucumber-engine) from Cucumber team.

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

    exitCode := s.Run()
    os.Exit(exitCode)
}
```

## TODO

* Support pending steps
* Formatters (pretty, dots)
* Filtering (single scenario)
* Support Cucumber expressions (currently regex)
* godoc

## License

This repository is licensed under the MIT license.
