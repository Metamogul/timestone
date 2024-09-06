package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newWaitGroups(t *testing.T) {
	t.Parallel()

	newWaitGroups := newWaitGroups()

	require.NotNil(t, newWaitGroups)
	require.Empty(t, newWaitGroups.waitGroups)
}

func Test_waitGroups_new(t *testing.T) {
	t.Parallel()

	w := newWaitGroups()

	w.new("test")
	require.Len(t, w.waitGroups, 1)
	require.NotNil(t, w.waitGroups["test"])
}

func Test_waitGroups_add(t *testing.T) {
	t.Parallel()

	t.Run("action name exists", func(t *testing.T) {
		t.Parallel()

		w := newWaitGroups()
		w.new("test")

		w.add("test", 1)
		go func() { w.done("test") }()
		w.waitFor("test")
	})

	t.Run("action name doesn't exist", func(t *testing.T) {
		t.Parallel()

		w := newWaitGroups()
		require.Panics(t, func() { w.add("notRegistered", 1) })
	})
}

func Test_waitGroups_done(t *testing.T) {
	t.Parallel()

	t.Run("action name exists", func(t *testing.T) {
		t.Parallel()

		w := newWaitGroups()
		w.new("test")

		w.add("test", 1)
		go func() { w.done("test") }()
		w.waitFor("test")
	})

	t.Run("action name doesn't exist", func(t *testing.T) {
		t.Parallel()

		w := newWaitGroups()
		require.Panics(t, func() { w.done("notRegistered") })
	})
}

func Test_waitGroups_waitFor(t *testing.T) {
	t.Parallel()

	t.Run("action name exists", func(t *testing.T) {
		t.Parallel()

		w := newWaitGroups()
		w.new("test1")
		w.new("test2")

		w.add("test1", 1)
		w.add("test2", 1)
		go func() {
			w.done("test1")
			w.done("test2")
		}()
		w.waitFor("test1", "test2")
	})

	t.Run("action name doesn't exist", func(t *testing.T) {
		t.Parallel()

		w := newWaitGroups()
		require.Panics(t, func() { w.waitFor("notRegistered") })
	})
}

func Test_waitGroups_wait(t *testing.T) {
	t.Parallel()

	w := newWaitGroups()

	for i := range 5 {
		testName := fmt.Sprintf("test%d", i)
		w.new(testName)
		w.add(testName, 1)
		go func() { w.done(testName) }()
	}

	w.wait()
}
