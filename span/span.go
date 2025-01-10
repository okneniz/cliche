package span

type Interface interface {
	From() int
	To() int
	Empty() bool
	Size() int
	Include(int) bool
	String() string
}
