package waitgroups

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewGeneratorWaitGroups(t *testing.T) {
	t.Parallel()

	newWaitGroups := NewGeneratorWaitGroups()

	require.NotNil(t, newWaitGroups)
	require.Empty(t, newWaitGroups.waitGroups.All())
}

func Test_GeneratorWaitGroups_Add(t *testing.T) {
	t.Parallel()

	w := NewGeneratorWaitGroups()

	w.Add(1, []string{"test1"})
	go func() { w.Done([]string{"test1", "test2"}) }()
	w.WaitFor([]string{"test1"})
}

func Test_GeneratorWaitGroups_Done(t *testing.T) {
	t.Parallel()

	w := NewGeneratorWaitGroups()

	t.Run("one exact matching call", func(t *testing.T) {
		w.Add(1, []string{"testGroup", "test1"})
		go func() {
			w.Done([]string{"testGroup", "test1"})
		}()
		w.WaitFor([]string{"testGroup", "test1"})
	})

	t.Run("sufficient done calls", func(t *testing.T) {
		w.Add(2, []string{"testGroup"})
		go func() {
			w.Done([]string{"testGroup", "test1"})
			w.Done([]string{"testGroup", "test2"})
		}()
		w.WaitFor([]string{"testGroup"})
	})

	t.Run("more than sufficient done calls", func(t *testing.T) {
		w.Add(2, []string{"testGroup"})
		go func() {
			w.Done([]string{"testGroup", "test1"})
			w.Done([]string{"testGroup", "test2"})
			w.Done([]string{"testGroup", "test3"})
		}()
		w.WaitFor([]string{"testGroup"})
	})

}

func Test_GeneratorWaitGroups_WaitFor(t *testing.T) {
	t.Parallel()

	// TODO: add more cases
	// - cover multiple matching waitgroups for tagset
	// - test behavior that enables waiting for unavailable wgs

	t.Run("wait group for tags exists", func(t *testing.T) {
		t.Parallel()

		w := NewGeneratorWaitGroups()

		w.Add(1, []string{"test1", "test2"})
		w.Add(1, []string{"test3", "test4"})
		go func() {
			w.Done([]string{"test1", "test2"})
			w.Done([]string{"test3", "test4"})
		}()
		w.WaitFor([]string{"test1", "test2"})
		w.WaitFor([]string{"test3", "test4"})
	})

	t.Run("wait group for tags doesn't exist", func(t *testing.T) {
		t.Parallel()

		w := NewGeneratorWaitGroups()
		require.Panics(t, func() { w.WaitFor([]string{"test5", "test2"}) })
	})
}
