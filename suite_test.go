package cucumber_test

import (
	"fmt"
	"testing"

	"github.com/pranas/cucumber-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	s, err := cucumber.NewSuite()
	require.NoError(t, err)

	s.DefineStep(`^you concat "([^"]*)" and "([^"]*)"$`, func(tc cucumber.TestCase, matches ...string) error {
		tc.Set("state", matches[0] + matches[1])
		return nil
	})
	s.DefineStep(`^you should have "([^"]*)"$`, func(tc cucumber.TestCase, expected ...string) error {
		actual := tc.Get("state").(string)

		if actual != expected[0] {
			return fmt.Errorf("expected %s but got %s", expected[0], actual)
		}

		return nil
	})
	summary := s.Run()
	assert.True(t, summary.Success)
	assert.Equal(t, 0, summary.ExitCode)
	assert.Equal(t, 2, summary.TestCasesTotal)
	assert.Equal(t, 2, summary.TestCasesPassed)
	assert.Equal(t, 4, summary.StepsTotal)
	assert.Equal(t, 4, summary.StepsPassed)
}
