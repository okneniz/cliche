package regular

import (
	"encoding/json"
	"fmt"
)

// https://www.regular-expressions.info/repeat.html

type OutOfBounds struct {
	Min   int
	Max   int
	Value int
}

func (err OutOfBounds) Error() string {
	return fmt.Sprintf("%d is ouf of bounds %d..%d", err.Value, err.Min, err.Max)
}

// bnf / ebnf
//
// https://www2.cs.sfu.ca/~cameron/Teaching/384/99-3/regexp-plg.html

// do a lot of methods for different scanning
// - for match without allocations
// - for replacements
// - for data extractions
//
// and scanner for all of them?
//
// try to copy official API
//
// https://pkg.go.dev/regexp#Regexp.FindString
//
// https://swtch.com/~rsc/regexp/regexp2.html#posix
//
// https://www.rfc-editor.org/rfc/rfc9485.html#name-multi-character-escapes

type TextBuffer interface {
	ReadAt(int) rune
	Size() int
	Substring(int, int) (string, error)
	String() string
}

// buffer for groups captures
// TODO : use pointers instead string for unnamed groups?
type captures struct {
	from  map[string]int
	to    map[string]int
	order []string
}

func newCaptures() *captures {
	return &captures{
		from:  make(map[string]int),
		to:    make(map[string]int),
		order: make([]string, 0),
	}
}

func (c *captures) String() string {
	x := make(map[string]interface{})
	x["from"] = c.from
	x["to"] = c.to
	x["order"] = c.order

	js, err := json.Marshal(x)
	if err != nil {
		return err.Error()
	}

	return string(js)
}

func (c *captures) IsEmpty() bool {
	return len(c.order) == 0
}

func (c *captures) Size() int {
	return len(c.order)
}

func (c *captures) From(name string, index int) {
	if _, exists := c.from[name]; exists {
		return
	}

	c.from[name] = index
	c.order = append(c.order, name)
}

func (c *captures) To(name string, index int) {
	if _, exists := c.from[name]; exists {
		c.to[name] = index
	}
}

func (c *captures) Delete(name string) {
	delete(c.from, name)
	delete(c.to, name)
	// TODO : maybe use map + slice for faster remove?
	c.order = remove[string](c.order, name)
}

// TODO : check defaultFinish must be optional or remove it?
func (c *captures) ToSlice(defaultFinish int) []bounds {
	result := make([]bounds, 0, len(c.to))

	var (
		start  int
		finish int
		exists bool
	)

	// fmt.Printf("captures to slice: %#v\n", c)

	for _, name := range c.order {
		if start, exists = c.from[name]; !exists {
			break // TODO : or continue?
		}

		if finish, exists = c.to[name]; !exists {
			finish = defaultFinish // TODO : remove it?
		}

		result = append(result, bounds{
			from: start,
			to:   finish,
		})
	}

	return result
}

func (c *captures) ToMap(defaultFinish int) map[string]bounds {
	result := make(map[string]bounds, len(c.to))

	var (
		start  int
		finish int
		exists bool
	)

	for _, name := range c.order {
		if start, exists = c.from[name]; !exists {
			break
		}

		if finish, exists = c.to[name]; !exists {
			finish = defaultFinish
		}

		result[name] = bounds{
			from: start,
			to:   finish,
		}
	}

	return result
}

func remove[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}

	return l
}

type bounds struct {
	from, to int
}

func (b bounds) String() string {
	return fmt.Sprintf("%d-%d", b.from, b.to)
}

type boundsList[T Match] struct {
	data map[int]T
	max  *T
}

func newBoundsList[T Match]() *boundsList[T] {
	b := new(boundsList[T])
	b.data = make(map[int]T)
	return b
}

func (b *boundsList[T]) clear() {
	b.data = make(map[int]T, len(b.data))
	b.max = nil
}

func (b *boundsList[T]) push(newMatch T) {
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

func (b *boundsList[T]) maximum() *T {
	return b.max
}

func (b *boundsList[T]) toMap() map[int]T {
	return b.data
}

func (b *boundsList[T]) longestMatch(x, y T) T {
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
