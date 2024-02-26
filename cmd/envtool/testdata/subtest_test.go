package testdata

import "testing"

func TestWithSubtest(t *testing.T) {
	t.Run("First", func(t *testing.T) {})
	t.Run("Second", func(t *testing.T) {})
	t.Run("Third", func(t *testing.T) {
		t.Run("NestedOne", func(t *testing.T) {})
		t.Run("NestedTwo", func(t *testing.T) {})
		t.Run("NestedThree", func(t *testing.T) {})
	})
}
