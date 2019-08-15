package cucumber

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSuite(t *testing.T) {
	s, err := NewSuite(Config{})
	assert.NoError(t, err)

	assert.Equal(t, "en", s.config.Language)
	assert.NotEqual(t, 0, s.config.Seed)
	assert.Equal(t, []string{"features/"}, s.config.Paths)

	s, err = NewSuite(Config{
		Seed:        uint64(321),
		Concurrency: uint64(10),
		Strict:      true,
	}, "--seed", "123", "-c", "1", "--fast", "--dry", "features/concat.feature")
	assert.NoError(t, err)
	assert.Equal(t, uint64(123), s.config.Seed)
	assert.Equal(t, []string{"features/concat.feature"}, s.config.Paths)
	assert.Equal(t, uint64(1), s.config.Concurrency)
	assert.True(t, s.config.FailFast)
	assert.True(t, s.config.Strict)
	assert.True(t, s.config.DryRun)

	_, err = NewSuite(Config{}, "feature/non_existing_file")
	if assert.Error(t, err) {
		assert.Equal(t, "failed to find features in path: feature/non_existing_file", err.Error())
	}

	_, err = NewSuite(Config{}, "--fasst")
	if assert.Error(t, err) {
		assert.Equal(t, "flag provided but not defined: -fasst", err.Error())
	}
}
