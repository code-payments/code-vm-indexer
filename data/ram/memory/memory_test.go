package memory

import (
	"testing"

	"github.com/code-payments/code-vm-indexer/data/ram/tests"
)

func TestRamMemoryStore(t *testing.T) {
	testStore := New()
	teardown := func() {
		testStore.(*store).reset()
	}
	tests.RunTests(t, testStore, teardown)
}
