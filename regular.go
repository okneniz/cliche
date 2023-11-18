package regular

import (
	"bytes"
	"encoding/json"
	"fmt"

	c "github.com/okneniz/parsec/common"
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

type Match interface {
	From() int
	To() int
	Size() int
	String() string
}

type TextBuffer interface {
	ReadAt(int) rune
	Size() int
	Substring(int, int) (string, error)
	String() string
}

type Handler interface { // TODO : should be generic type for different type of matches?
	// String() string // TODO : implement it for debug

	Match(n node, from, to int, isLeaf, isEmpty bool)

	FirstMatch() *match
	FirstNotEmptyMatch() *match

	// TODO : how to remove it?
	// required only for quantifier
	LastMatch() *match // TODO : use (int, int) instead?
	LastNotEmptyMatch() *match

	Position() int
	Rewind(size int)

	AddNamedGroup(name string, index int)
	MatchNamedGroup(name string, index int)
	DeleteNamedGroup(name string)

	AddGroup(name string, index int)
	MatchGroup(name string, index int)
	DeleteGroup(name string)
}

type Callback func(x node, from int, to int, empty bool)

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

type match struct {
	from  int
	to    int
	node  node
	empty bool
}

func (m match) From() int {
	return m.from
}

func (m match) To() int {
	return m.to
}

func (m match) Empty() bool {
	return m.empty
}

func (m match) String() string {
	return fmt.Sprintf("%s - [%d..%d] %v", m.node.getKey(), m.from, m.to, m.empty)
}

func (m match) Size() int {
	if m.empty {
		return 0
	}

	return m.to - m.from + 1
}

type Trie interface {
	Add(...string) error
	Size() int
	MarshalJSON() ([]byte, error)
	String() string
	Match(string) []*FullMatch
}

var _ Trie = new(trie)

type trie struct {
	nodes index
}

func NewTrie(regexps ...string) (*trie, error) {
	tr := new(trie)
	tr.nodes = make(index)

	for _, regexp := range regexps {
		err := tr.Add(regexp)
		if err != nil {
			return nil, err
		}
	}

	return tr, nil
}

func (t *trie) Add(strs ...string) error {
	for _, str := range strs {
		buf := newBuffer(str)

		node, err := defaultParser(buf)
		if err != nil {
			return err
		}

		t.addExpression(str, node)
	}

	return nil
}

func (t *trie) Size() int {
	size := 0

	for _, x := range t.nodes {
		x.walk(func(n node) {
			size++
		})
	}

	return size
}

func (t *trie) MarshalJSON() ([]byte, error) {
	scanner := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(scanner)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	err := encoder.Encode(t.nodes)
	if err != nil {
		return nil, err
	}

	return scanner.Bytes(), nil
}

func (t *trie) String() string {
	data, err := t.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(data)
}

func (t *trie) addExpression(str string, newNode node) {
	newNode.walk(func(x node) {
		if len(x.getNestedNodes()) == 0 {
			x.addExpression(str)
		}
	})

	key := newNode.getKey()

	if prev, exists := t.nodes[key]; exists {
		prev.merge(newNode)
	} else {
		t.nodes[key] = newNode
	}
}

func (t *trie) Match(text string) []*FullMatch {
	if len(text) == 0 {
		return nil
	}

	input := newBuffer(text)
	matches := newBoundsList[match]()
	groups := newCaptures() // TODO : use node as key for unnamed groups to avoid generate string ID
	namedGroups := newCaptures()

	// DS for best matches - https://web.engr.oregonstate.edu/~erwig/diet/
	acc := make(map[node]*boundsList[*FullMatch])

	var scanner *fullScanner

	scanner = newFullScanner(
		groups,
		namedGroups,
		func(n node, from, to int, empty bool) {
			matches.push(
				match{
					from:  from,
					to:    to,
					node:  n,
					empty: empty,
				},
			)

			begin := scanner.FirstMatch()
			end := scanner.LastMatch()

			beginSubstring := scanner.FirstNotEmptyMatch()
			endSubstring := scanner.LastNotEmptyMatch()

			fmt.Println("scanner", scanner)

			m := &FullMatch{
				expressions: n.getExpressions().toSlice(),
				from:        begin.From(),
				to:          end.To(),
				groups:      groups.ToSlice(0), // TODO : default start?
				namedGroups: namedGroups.ToMap(0),
			}

			if m.from >= input.Size() {
				m.from = input.Size() - 1
			}

			if m.to >= input.Size() {
				m.to = input.Size() - 1
			}

			if beginSubstring != nil && endSubstring != nil {
				subString, err := input.Substring(
					beginSubstring.From(),
					endSubstring.To(),
				)

				if err != nil {
					// TODO : how to handle error
					fmt.Println("error", err)
				}

				m.subString = subString
			} else {
				m.empty = true
			}

			fmt.Printf("full match: %v\n", m)

			if list, exists := acc[n]; exists {
				list.push(m)
			} else {
				newList := newBoundsList[*FullMatch]()
				newList.push(m)
				acc[n] = newList
			}

			fmt.Println(" ")
		},
	)

	from := 0
	to := input.Size() - 1

	// - как правильно матчить
	// - как избегать лишних сканирований?

	for _, n := range t.nodes {
		nextFrom := from

		for nextFrom <= to {
			n.match(scanner, input, nextFrom, to, func(x node, f, t int, _ bool) {
				// if n.isEnd() {
				// 	fmt.Printf("match %v '%s' from %d to %d\n", x.getExpressions(), x.getKey(), nextFrom, nextTo)
				// }
			})

			longestMatch := matches.maximum() // maybe rename to best?

			if longestMatch != nil {
				nextFrom = longestMatch.To() + 1
			} else {
				nextFrom += 1
			}

			scanner.Rewind(0)
			matches.clear()
		}
	}

	result := make([]*FullMatch, 0, len(acc))
	for _, list := range acc {
		for _, item := range list.toMap() {
			result = append(result, item)
		}
	}

	return result
}

type simpleBuffer struct {
	data     []rune
	position int

	// data string
	// positions []int (stack of last positions)
}

// newBuffer - make buffer which can read text on input

var _ c.Buffer[rune, int] = &simpleBuffer{}

func newBuffer(str string) *simpleBuffer {
	b := new(simpleBuffer)
	b.data = []rune(str)
	b.position = 0
	return b
}

// Read - read next item, if greedy buffer keep position after reading.
func (b *simpleBuffer) Read(greedy bool) (rune, error) {
	if b.IsEOF() {
		return 0, c.EndOfFile
	}

	x := b.data[b.position]

	if greedy {
		b.position++
	}

	return x, nil
}

func (b *simpleBuffer) ReadAt(idx int) rune {
	return b.data[idx]
}

func (b *simpleBuffer) Size() int { // TODO : check for another runes
	return len(b.data)
}

func (b *simpleBuffer) String() string {
	return fmt.Sprintf("Buffer(%s, %d)", string(b.data), b.position)
}

func (b *simpleBuffer) Substring(from, to int) (string, error) {
	if from > to {
		return "", fmt.Errorf(
			"invalid bounds for substring: from=%d to=%d size=%d",
			from,
			to,
			len(b.data),
		)
	}

	if from < 0 || from >= len(b.data) || to >= len(b.data) {
		return "", fmt.Errorf(
			"out of bounds buffer: from=%d to=%d size=%d",
			from,
			to,
			len(b.data),
		)
	}

	return string(b.data[from : to+1]), nil
}

// Seek - change buffer position
func (b *simpleBuffer) Seek(x int) {
	b.position = x
}

// Position - return current buffer position
func (b *simpleBuffer) Position() int {
	return b.position
}

// IsEOF - true if buffer ended
func (b *simpleBuffer) IsEOF() bool {
	return b.position >= len(b.data)
}

type index map[string]node

func (ix index) merge(other index) {
	for key, newNode := range other {
		if prev, exists := ix[key]; exists {
			prev.merge(newNode)
		} else {
			ix[key] = newNode
		}
	}
}

type dict map[string]struct{}

func (d dict) add(str string) {
	d[str] = struct{}{}
}

func (d dict) merge(other dict) {
	for key, value := range other {
		d[key] = value
	}
}

func (d dict) toSlice() []string {
	result := make([]string, len(d))
	i := 0
	for key := range d {
		result[i] = key
		i++
	}
	return result
}
