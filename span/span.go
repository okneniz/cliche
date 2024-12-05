package span

type Interface interface {
	From() int
	To() int
	Empty() bool
	Size() int
	IsInclude(int) bool
	String() string
}
