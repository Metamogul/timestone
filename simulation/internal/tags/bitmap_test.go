package tags

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_newBitmap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		index int
		want  bitmap
	}{
		{
			name:  "index 0",
			index: 0,
			want:  bitmap{1 << 0},
		},
		{
			name:  "index 63",
			index: 63,
			want:  bitmap{1 << 63},
		},
		{
			name:  "index 64",
			index: 64,
			want:  bitmap{0, 1 << 0},
		},
		{
			name:  "index 127",
			index: 127,
			want:  bitmap{0, 1 << 63},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newBitmap(tt.index)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_bitmap_or(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		b    bitmap
		bb   bitmap
		want bitmap
	}{
		{
			name: "only empty",
			b:    bitmap{},
			bb:   bitmap{},
			want: bitmap{},
		},
		{
			name: "identity with empty bitmap",
			b:    bitmap{1 << 63},
			bb:   bitmap{},
			want: bitmap{1 << 63},
		},
		{
			name: "identity with zero bitmap",
			b:    bitmap{1 << 63},
			bb:   bitmap{},
			want: bitmap{1 << 63},
		},
		{
			name: "b OR bb, equal length",
			b:    bitmap{0, 1 << 1},
			bb:   bitmap{1 << 1, 1 << 2},
			want: bitmap{0 | 1<<1, 1<<1 | 1<<2},
		},
		{
			name: "b OR bb, b longer",
			b:    bitmap{0, 1 << 1, 0},
			bb:   bitmap{1 << 1, 1 << 2},
			want: bitmap{0 | 1<<1, 1<<1 | 1<<2, 0},
		},
		{
			name: "b OR bb, bb longer",
			b:    bitmap{0, 1 << 1},
			bb:   bitmap{1 << 1, 1 << 2, 0},
			want: bitmap{0 | 1<<1, 1<<1 | 1<<2, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.b.or(tt.bb)
			require.Equal(t, tt.want, tt.b)
		})
	}
}

func Test_bitmap_contains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		b      bitmap
		target bitmap
		want   bool
	}{
		{
			name:   "zero target",
			b:      bitmap{1 << 0},
			target: bitmap{0},
			want:   true,
		},
		{
			name:   "contains target",
			b:      bitmap{1<<1 | 1<<0, 1 << 0},
			target: bitmap{1 << 0},
			want:   true,
		},
		{
			name:   "equals target",
			b:      bitmap{1<<1 | 1<<0},
			target: bitmap{1<<1 | 1<<0},
			want:   true,
		},
		{
			name:   "does not contain target",
			b:      bitmap{1<<1 | 1<<0},
			target: bitmap{1 << 2},
			want:   false,
		},
		{
			name:   "target too large",
			b:      bitmap{1<<1 | 1<<0},
			target: bitmap{1 << 2, 1 << 0},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.b.contains(tt.target)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_bitmap_containedIn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		b      bitmap
		target bitmap
		want   bool
	}{
		{
			name:   "zero bitmap",
			b:      bitmap{0},
			target: bitmap{1 << 0},
			want:   true,
		},
		{
			name:   "contained in target",
			b:      bitmap{1 << 0},
			target: bitmap{1<<1 | 1<<0, 1 << 0},
			want:   true,
		},
		{
			name:   "equals target",
			b:      bitmap{1<<1 | 1<<0},
			target: bitmap{1<<1 | 1<<0},
			want:   true,
		},
		{
			name:   "is not contained in target",
			b:      bitmap{1 << 2},
			target: bitmap{1<<1 | 1<<0},
			want:   false,
		},
		{
			name:   "bitmap too large",
			b:      bitmap{1 << 2, 1 << 0},
			target: bitmap{1<<1 | 1<<0},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.b.containedIn(tt.target)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_bitmap_equal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		b      bitmap
		target bitmap
		want   bool
	}{
		{
			name:   "empty bitmaps",
			b:      bitmap{},
			target: bitmap{},
			want:   true,
		},
		{
			name:   "zero bitmaps",
			b:      bitmap{0},
			target: bitmap{0},
			want:   true,
		},
		{
			name:   "equal bitmaps",
			b:      bitmap{1 << 1},
			target: bitmap{1 << 1},
			want:   true,
		},
		{
			name:   "equal length, unequal bitmaps",
			b:      bitmap{1 << 1},
			target: bitmap{1 << 2},
			want:   false,
		},
		{
			name:   "unequal lengths",
			b:      bitmap{1 << 1},
			target: bitmap{1<<0 | 1<<2},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.b.equal(tt.target)
			require.Equal(t, tt.want, got)
		})
	}
}
