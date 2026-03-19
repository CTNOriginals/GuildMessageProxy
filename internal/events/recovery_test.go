package events

import (
	"testing"
)

func TestRecoverPanic(t *testing.T) {
	// Test that recoverPanic catches panics
	defer func() {
		if r := recover(); r != nil {
			t.Error("recoverPanic should have caught the panic")
		}
	}()

	defer recoverPanic("test")
	panic("test panic")
}
