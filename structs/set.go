package structs

type Set[T comparable] interface {
	Add(item ...T)
	AddTo(other Set[T])
	Size() int
	Slice() []T
	Clone() Set[T]
}

type mapSet[T comparable] struct {
	items map[T]struct{}
}

var _ Set[string] = NewMapSet[string]()

func NewMapSet[T comparable](items ...T) *mapSet[T] {
	s := &mapSet[T]{
		items: make(map[T]struct{}, len(items)),
	}

	s.Add(items...)

	return s
}

func (s *mapSet[T]) Add(items ...T) {
	for _, item := range items {
		s.items[item] = struct{}{}
	}
}

func (s *mapSet[T]) AddTo(other Set[T]) {
	for item, _ := range s.items {
		other.Add(item)
	}
}

func (s *mapSet[T]) Slice() []T {
	list := make([]T, len(s.items))

	i := 0
	for key := range s.items {
		list[i] = key
		i++
	}

	return list
}

func (s *mapSet[T]) Size() int {
	return len(s.items)
}

func (s *mapSet[T]) Clone() Set[T] {
	newSet := &mapSet[T]{
		items: make(map[T]struct{}, len(s.items)),
	}

	s.AddTo(newSet)

	return newSet
}
