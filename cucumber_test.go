package cucumber_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/pranas/cucumber-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	summary := cucumber.NewSummaryFormatter(ioutil.Discard)
	s, err := cucumber.NewSuite(cucumber.Config{Formatter:summary})
	require.NoError(t, err)

	exitCode := s.Run()
	assert.Equal(t, 1, exitCode)
	assert.False(t, summary.Success)
	assert.Equal(t, 2, summary.TestCasesTotal)
	assert.Equal(t, 0, summary.TestCasesPassed)
	assert.Equal(t, 4, summary.StepsTotal)
	assert.Equal(t, 0, summary.StepsPassed)

	summary = cucumber.NewSummaryFormatter(ioutil.Discard)
	s, err = cucumber.NewSuite(cucumber.Config{Formatter:summary})
	require.NoError(t, err)

	s.DefineStep(`^you concat "([^"]*)" and "([^"]*)"$`, concat)
	s.DefineStep(`^you should have "([^"]*)"$`, matchOutput)

	exitCode = s.Run()
	assert.Equal(t, 0, exitCode)
	assert.True(t, summary.Success)
	assert.Equal(t, 2, summary.TestCasesTotal)
	assert.Equal(t, 2, summary.TestCasesPassed)
	assert.Equal(t, 4, summary.StepsTotal)
	assert.Equal(t, 4, summary.StepsPassed)

	summary = cucumber.NewSummaryFormatter(ioutil.Discard)
	s, err = cucumber.NewSuite(cucumber.Config{Formatter:summary}, "features/concat.feature:6")
	require.NoError(t, err)

	exitCode = s.Run()
	assert.Equal(t, 1, exitCode)
	assert.False(t, summary.Success)
	assert.Equal(t, 1, summary.TestCasesTotal)
	assert.Equal(t, 0, summary.TestCasesPassed)
	assert.Equal(t, 2, summary.StepsTotal)
	assert.Equal(t, 0, summary.StepsPassed)

	summary = cucumber.NewSummaryFormatter(ioutil.Discard)
	s, err = cucumber.NewSuite(cucumber.Config{Formatter:summary})
	require.NoError(t, err)

	s.DefineStep(`^you concat "([^"]*)" and "([^"]*)"$`, concat)
	s.DefineStep(`^you should have "([^"]*)"$`, func(_ cucumber.TestCase, s ...string) error {
		panic("should recover from this")
	})

	exitCode = s.Run()
	assert.Equal(t, 1, exitCode)
	assert.False(t, summary.Success)
	assert.Equal(t, 2, summary.TestCasesTotal)
	assert.Equal(t, 0, summary.TestCasesPassed)
	assert.Equal(t, 4, summary.StepsTotal)
	assert.Equal(t, 2, summary.StepsPassed)
}

func concat(tc cucumber.TestCase, matches ...string) error {
	tc.Set("state", matches[0] + matches[1])
	return nil
}

func matchOutput(tc cucumber.TestCase, expected ...string) error {
	actual := tc.Get("state").(string)

	if actual != expected[0] {
		return fmt.Errorf("expected %s but got %s", expected[0], actual)
	}

	return nil
}