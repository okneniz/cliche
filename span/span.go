package span

type Interface interface {
	From() int
	To() int
	Empty() bool
	Size() int
	Include(int) bool
	String() string
}

func New(from, to int, empty bool) Interface {
	if empty {
		return Empty(from)
	}

	return Pair(from, to)
}

func Compare(m1, m2 Interface) int {
	switch {
	case m1.From() > m2.From():
		return -1
	case m1.From() < m2.From():
		return 1
	default:
		switch {
		case m1.Size() > m2.Size():
			return 1
		case m1.Size() < m2.Size():
			return -1
		default:
			return 0
		}
	}
}

func NextFrom(pos int, empty bool) int {
	if empty {
		return pos
	}

	return pos + 1
}
