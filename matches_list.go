package regular

type matchesList[T Match] struct {
	data map[int]T
	max  *T
}

func newMatchesList[T Match]() *matchesList[T] {
	b := new(matchesList[T])
	b.data = make(map[int]T)
	return b
}

func (b *matchesList[T]) clear() {
	b.data = make(map[int]T, len(b.data))
	b.max = nil
}

func (b *matchesList[T]) push(newMatch T) {
	if prevMatch, exists := b.data[newMatch.From()]; exists {
		b.data[newMatch.From()] = b.longestMatch(prevMatch, newMatch)
	} else {
		b.data[newMatch.From()] = newMatch
	}

	if b.max == nil {
		b.max = &newMatch
	} else {
		newMax := b.longestMatch(*b.max, newMatch)
		b.max = &newMax
	}
}

func (b *matchesList[T]) maximum() *T {
	return b.max
}

func (b *matchesList[T]) toMap() map[int]T {
	return b.data
}

func (b *matchesList[T]) longestMatch(x, y T) T {
	if x.Size() == y.Size() {
		if x.From() < y.From() {
			return x
		}

		return y
	}

	if x.Size() > y.Size() {
		return x
	}

	return y
}
