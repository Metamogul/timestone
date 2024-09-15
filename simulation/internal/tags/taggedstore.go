package tags

// To support very large sets, there's great potential
// for optimization here using a prefix tree as an index,
// as well as compressing sparse bitmaps. For small sets
// of a few ten to a few thousand items, this is sufficiently
// good.

type taggedValue[T any] struct {
	bitmap
	value T
}

type TaggedStore[T any] struct {
	bitmapsByTags map[string]bitmap
	content       []taggedValue[T]
}

func NewTaggedStore[T any]() *TaggedStore[T] {
	return &TaggedStore[T]{
		bitmapsByTags: make(map[string]bitmap),
		content:       make([]taggedValue[T], 0),
	}
}

func (t *TaggedStore[T]) Set(value T, tags []string) {
	if len(tags) == 0 {
		panic("tags must not be empty")
	}

	bitmapForTags := t.bitmapForTags(tags)

	// If entry exists, replace value
	for i, entry := range t.content {
		if entry.bitmap.equal(bitmapForTags) {
			t.content[i].value = value
			return
		}
	}

	// Create new entry
	t.content = append(t.content, taggedValue[T]{
		bitmap: bitmapForTags,
		value:  value,
	})
}

func (t *TaggedStore[T]) Containing(tags []string) []T {
	bitmaskForTags := t.bitmapForTags(tags)

	result := make([]T, 0, len(t.content))
	for _, entry := range t.content {
		if entry.contains(bitmaskForTags) {
			result = append(result, entry.value)
		}
	}

	return result
}

func (t *TaggedStore[T]) ContainedIn(tags []string) []T {
	bitmaskForTags := t.bitmapForTags(tags)

	result := make([]T, 0, len(t.content))
	for _, entry := range t.content {
		if entry.containedIn(bitmaskForTags) {
			result = append(result, entry.value)
		}
	}

	return result
}

func (t *TaggedStore[T]) Matching(tags []string) T {
	bitmaskForTags := t.bitmapForTags(tags)

	for _, entry := range t.content {
		if entry.equal(bitmaskForTags) {
			return entry.value
		}
	}

	return *new(T)
}

func (t *TaggedStore[T]) All() []T {
	var result []T
	for _, entry := range t.content {
		result = append(result, entry.value)
	}

	return result
}

func (t *TaggedStore[T]) bitmapForTag(tag string) bitmap {
	bitmapForTag, ok := t.bitmapsByTags[tag]
	if !ok {
		bitmapForTag = newBitmap(len(t.bitmapsByTags))
		t.bitmapsByTags[tag] = bitmapForTag
	}

	return bitmapForTag
}

func (t *TaggedStore[T]) bitmapForTags(tags []string) bitmap {
	tagsBitmask := make(bitmap, len(t.bitmapsByTags)/64+1)

	for _, tag := range tags {
		tagsBitmask.or(t.bitmapForTag(tag))
	}

	return tagsBitmask
}
