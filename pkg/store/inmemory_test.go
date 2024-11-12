//go:build unit

package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_InMemoryKVStore_Put(t *testing.T) {
	t.Run("Puts the element in the store", func(t *testing.T) {
		// Arrange
		s := NewInMemoryKVStore[string, int]()

		key, val := "foo", 10

		// Act
		s.Put(key, val)

		// Assert
		retrieved, ok := s.data[key]

		assert.True(t, ok)
		assert.Equal(t, val, retrieved)
	})

	t.Run("Overwrites element in the store", func(t *testing.T) {
		// Arrange
		s := NewInMemoryKVStore[string, int]()

		key, val := "foo", 10
		s.data[key] = val - 1

		// Act
		s.Put(key, val)

		// Assert
		retrieved, ok := s.data[key]

		assert.True(t, ok)
		assert.Equal(t, val, retrieved)
	})
}

func Test_InMemoryKVStore_Get(t *testing.T) {
	t.Run("Retrieves the element from the store if it exists", func(t *testing.T) {
		// Arrange
		s := NewInMemoryKVStore[string, int]()
		key, val := "foo", 10

		s.data[key] = val

		// Act
		retrieved, ok := s.Get(key)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, val, retrieved)
	})

	t.Run("Returns false if key does not exist in teh store", func(t *testing.T) {
		// Arrange
		s := NewInMemoryKVStore[string, int]()
		key := "foo"

		// Act
		retrieved, ok := s.Get(key)

		// Assert
		assert.False(t, ok)
		assert.Equal(t, 0, retrieved)
	})
}
