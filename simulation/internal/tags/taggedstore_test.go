package tags

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Benchmark_TaggedStoreBitmask_GetContaining(b *testing.B) {
	ts := NewTaggedStore[string]()

	// Add values with tags
	ts.Set("apple", []string{"fruit", "red", "round", "borra", "bazza", "bumma", "climb", "result", "president"})
	ts.Set("banana", []string{"fruit", "yellow", "long", "borrb", "bazzb", "bummb"})
	ts.Set("carrot", []string{"vegetable", "orange", "long", "borrc", "bazzc", "bummc"})

	for n := 0; n < b.N; n++ {
		_ = ts.Containing([]string{"fruit", "red", "borra", "bazza", "bumma", "climb", "result", "president"})
		_ = ts.Containing([]string{"fruit", "long"})
		_ = ts.Containing([]string{"vegetable", "orange", "long"})
		_ = ts.Containing([]string{"red", "fruit"})
	}
}

func Test_NewTaggedStore(t *testing.T) {
	t.Parallel()

	ts := NewTaggedStore[string]()

	require.NotNil(t, ts)
	require.NotNil(t, ts.content)
	require.Empty(t, ts.content)
	require.NotNil(t, ts.bitmapsByTags)
	require.Empty(t, ts.bitmapsByTags)
}

func Test_TaggedStore_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		tags      []string
		wantPanic bool
	}{
		{
			name:      "no tags passed",
			value:     "value",
			wantPanic: true,
		},
		{
			name:      "success",
			value:     "value",
			tags:      []string{"foo", "bar", "baz"},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := NewTaggedStore[string]()

			if tt.wantPanic {
				require.Panics(t, func() {
					ts.Set(tt.value, tt.tags)
				})
				return
			}

			ts.Set(tt.value, tt.tags)
			for _, tag := range tt.tags {
				result := ts.Containing([]string{tag})
				require.NotNil(t, result)
				require.Equal(t, result[0], tt.value)
			}
		})
	}
}

func Test_TaggedStore_Containing(t *testing.T) {
	t.Parallel()

	value1 := "value"
	tags1 := []string{"foo", "bar", "baz"}

	value2 := "apple"
	tags2 := []string{"fruit", "round", "red", "foo"}

	tests := []struct {
		name       string
		getForTags []string
		want       []string
	}{
		{
			name:       "no matches",
			getForTags: []string{"bum"},
			want:       []string{},
		},
		{
			name:       "match value1",
			getForTags: []string{"foo", "bar"},
			want:       []string{value1},
		},
		{
			name:       "match value1 and value2",
			getForTags: []string{"foo"},
			want:       []string{value1, value2},
		},
		{
			name:       "match value2",
			getForTags: []string{"fruit"},
			want:       []string{value2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := NewTaggedStore[string]()

			ts.Set(value1, tags1)
			ts.Set(value2, tags2)

			value := ts.Containing(tt.getForTags)
			require.Equal(t, tt.want, value)
		})
	}
}

func Test_TaggedStore_ContainedIn(t *testing.T) {
	t.Parallel()

	value1 := "value"
	tags1 := []string{"foo", "bar"}

	value2 := "apple"
	tags2 := []string{"fruit", "round"}

	tests := []struct {
		name       string
		getForTags []string
		want       []string
	}{
		{
			name:       "no matches",
			getForTags: []string{"foo", "bum", "baz"},
			want:       []string{},
		},
		{
			name:       "match value1",
			getForTags: []string{"foo", "bar", "baz"},
			want:       []string{value1},
		},
		{
			name:       "match value1 and value2",
			getForTags: []string{"foo", "bar", "baz", "fruit", "round", "red"},
			want:       []string{value1, value2},
		},
		{
			name:       "match value2",
			getForTags: []string{"fruit", "round", "red"},
			want:       []string{value2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := NewTaggedStore[string]()

			ts.Set(value1, tags1)
			ts.Set(value2, tags2)

			value := ts.ContainedIn(tt.getForTags)
			require.Equal(t, tt.want, value)
		})
	}
}

func Test_TaggedStore_Matching(t *testing.T) {
	t.Parallel()

	value1 := "value"
	tags1 := []string{"foo", "bar", "baz"}

	value2 := "apple"
	tags2 := []string{"fruit", "round", "red", "foo"}

	tests := []struct {
		name       string
		getForTags []string
		want       string
	}{
		{
			name:       "no matches",
			getForTags: []string{"bum"},
			want:       "",
		},
		{
			name:       "match value1",
			getForTags: []string{"foo", "bar", "baz"},
			want:       value1,
		},
		{
			name:       "don't match subsets",
			getForTags: []string{"foo", "bar"},
			want:       "",
		},
		{
			name:       "match value2",
			getForTags: []string{"fruit", "round", "red", "foo"},
			want:       value2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := NewTaggedStore[string]()

			ts.Set(value1, tags1)
			ts.Set(value2, tags2)

			value := ts.Matching(tt.getForTags)
			require.Equal(t, tt.want, value)
		})
	}
}

func Test_TaggedStore_All(t *testing.T) {
	t.Parallel()

	value1 := "value"
	tags1 := []string{"foo", "bar", "baz"}

	value2 := "apple"
	tags2 := []string{"fruit", "round", "red", "foo"}

	ts := NewTaggedStore[string]()

	ts.Set(value1, tags1)
	ts.Set(value2, tags2)

	values := ts.All()
	require.Equal(t, []string{value1, value2}, values)
}

func Test_TaggedStore_bitmapForTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		tagsAlreadyAdded []string
		tag              string
		want             bitmap
	}{
		{
			name:             "first tag",
			tagsAlreadyAdded: []string{},
			tag:              "foo",
			want:             bitmap{1 << 0},
		},
		{
			name:             "second tag",
			tagsAlreadyAdded: []string{"foo"},
			tag:              "bar",
			want:             bitmap{1 << 1},
		},
		{
			name:             "tag was added before",
			tagsAlreadyAdded: []string{"foo", "bar"},
			tag:              "bar",
			want:             bitmap{1 << 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := NewTaggedStore[string]()
			for _, tag := range tt.tagsAlreadyAdded {
				_ = ts.bitmapForTag(tag)
			}

			got := ts.bitmapForTag(tt.tag)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_TaggedStore_bitmapForTags(t *testing.T) {
	t.Parallel()

	tagsAlreadyAdded := []string{"foo", "bar", "baz"}

	tests := []struct {
		name string

		tags []string
		want bitmap
	}{
		{
			name: "one tag",
			tags: []string{"foo"},
			want: bitmap{1 << 0},
		},
		{
			name: "multiple tags",
			tags: []string{"foo", "baz"},
			want: bitmap{1<<0 | 1<<2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := NewTaggedStore[string]()
			for _, tag := range tagsAlreadyAdded {
				_ = ts.bitmapForTag(tag)
			}

			got := ts.bitmapForTags(tt.tags)
			require.Equal(t, tt.want, got)
		})
	}
}
