package load

import (
	"math"
	"sort"
)

type Tag struct {
	S string
	F float32
}
type tagSlice []Tag

func (t tagSlice) Less(a, b int) bool { return t[a].F < t[b].F }
func (t tagSlice) Len() int           { return len(t) }
func (t tagSlice) Swap(a, b int)      { t[a], t[b] = t[b], t[a] }

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (t tagSlice) Name(i int) string   { return t[i].S }
func (t tagSlice) Value(i int) float32 { return t[i].F }

// Calculates the most used tags.
//
// Returns a slice of [0..max] tags, sorted descending by popularity.
func (l *Loader) TopTags(max int) []string {
	tags := l.TagsFloat()
	ret := make([]string, min(max, len(tags)))
	for i := 0; i < len(ret); i++ {
		ret[i] = tags[i].S
	}
	return ret
}

// Returns tags, sorted by popularity
func (l *Loader) TagsFloat() tagSlice {
	tagsMap := l.TagsFloatMap()
	tags := make(tagSlice, 0, len(tagsMap))
	for s, f := range tagsMap {
		tags = append(tags, Tag{s, f})
	}
	sort.Sort(sort.Reverse(&tags))

	return tags
}

// Returns all tags, with float popularity in the range [0..1]
func (l *Loader) TagsFloatMap() map[string]float32 {
	tags := make(map[string]float32)

	// add everything up
	for _, t := range l.Tracks {
		for tag, count := range t.Tags {
			tags[tag] += float32(count)
		}
	}

	// find min, max
	var max, min float32
	min = math.MaxFloat32
	for _, c := range tags {
		if c > max {
			max = c
		}
		if c < min {
			min = c
		}
	}

	// normalize to [0..1]
	for t, c := range tags {
		tags[t] = (c - min) / (max - min)
	}
	return tags
}
