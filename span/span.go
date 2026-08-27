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
