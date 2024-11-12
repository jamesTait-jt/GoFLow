//go:build unit

package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Contains(t *testing.T) {
	t.Run("Returns true if slice contains the item", func(t *testing.T) {
		// Arrange
		n := 5
		ns := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		// Act
		ok := Contains(ns, n)

		// Assert
		assert.True(t, ok)
	})

	t.Run("Returns false if slice does not contain the item", func(t *testing.T) {
		// Arrange
		n := 500
		ns := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		// Act
		ok := Contains(ns, n)

		// Assert
		assert.False(t, ok)
	})
}
