package app

import (
	"fmt"
	"testing"
)

func TestTry(t *testing.T) {
	t.Run("panics when given a non-nil error", func(t *testing.T) {
		defer expectPanic(t, "error given to Try()")
		Try(fmt.Errorf("error given to Try()"))
	})
	t.Run("panics when given a non-nil error", func(t *testing.T) {
		Try(nil)
		t.Log("Try() called with nil error so no panic created")
	})
}
