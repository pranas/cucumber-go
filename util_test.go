package cucumber

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFeatures(t *testing.T) {
	files, err := findFeatures(".")
	assert.NoError(t, err)
	assert.Empty(t, files)

	files, err = findFeatures("features")
	assert.NoError(t, err)
	assert.Equal(t, []string{"features/concat.feature"}, files)

	files, err = findFeatures("features/concat.feature")
	assert.NoError(t, err)
	assert.Equal(t, []string{"features/concat.feature"}, files)

}
