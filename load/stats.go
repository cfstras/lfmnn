package load

// Calculates the most used tags.
//
// Returns a slice of [0..max] tags, sorted descending by popularity.
func (l *Loader) TopTags(max int) []string {
	tags := make(map[string]float32)
	for _, t := range l.Tracks {
		for tag, count := range t.Tags {
			//TODO calculate max, avg etc
			tags[tag] += count
		}
	}
}
