package regular

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	c "github.com/okneniz/parsec/common"
)

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

type Match interface {
	From() int
	To() int
	String() string
	NamedGroups() map[string]string
	Groups() []string
}

type TextReader interface {
	ReadAt(int) (rune, error)
}

type Handler interface {
	Input() TextReader
	Position() int
	FirstMatch() (int, int)
	LastMatch() (int, int)

	Rewind(size int)
	Match(n node, from, to int, isLeaf, isEmpty bool)

	StartNamedGroup(name string, index int)
	EndNamedGroup(name string, index int)
	DeleteNamedGroup(name string)

	StartGroup(name string, index int)
	EndGroup(name string, index int)
	DeleteGroup(name string)
}

type Callback func(x node, from int, to int)

// buffer for groups captures
type captures struct {
	from  map[string]int
	to    map[string]int
	order []string
}

func newCaptures() captures {
	return captures{
		from: make(map[string]int),
		to: make(map[string]int),
		order: make([]string, 0),
	}
}

func (c captures) IsEmpty() bool {
	return len(c.order) == 0
}

func (c captures) Size() int {
	return len(c.order)
}

func (c captures) From(name string, index int) {
	if _, exists := c.from[name]; exists {
		return
	}

	c.from[name] = index
	c.order = append(c.order, name)
}

func (c captures) To(name string, index int) {
	if _, exists := c.from[name]; exists {
		c.to[name] = index
	}
}

func (c captures) Delete(name string) {
	delete(c.from, name)
	delete(c.to, name)
	// TODO : maybe use map + slice for faster remove?
	c.order = remove[string](c.order, name)
}

// TODO : check defaultFinish must be optional?
func (c captures) ToSlice(defaultFinish int) []bounds {
	result := make([]bounds, 0, len(c.to))

	var (
		start  int
		finish int
		exists bool
	)

	for _, name := range c.order {
		if _, exists := c.from[name]; !exists {
			break
		}

		if finish, exists = c.to[name]; !exists {
			finish = defaultFinish
		}

		result = append(result, bounds{
			from: start,
			to:   finish,
		})
	}

	return result
}

// TODO : rename method to lower case?
func (c captures) ToMap(defaultFinish int) map[string]bounds {
	result := make(map[string]bounds, len(c.to))

	var (
		start  int
		finish int
		exists bool
	)

	for _, name := range c.order {
		if _, exists := c.from[name]; !exists {
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

type list[T any] struct {
	data []T
	pos  int
}

func newList[T any](cap int) *list[T] {
	l := new(list[T])
	l.data = make([]T, 0, cap)
	return l
}

func (l list[T]) append(item T) {
	if l.pos >= len(l.data) {
		l.data = append(l.data, item)
	} else {
		l.data[l.pos] = item
	}

	l.pos++
}

func (l list[T]) size() int {
	return len(l.data)
}

func (l list[T]) truncate(pos int) {
	if pos <= 0 {
		panic(fmt.Sprintf("invalid position for truncate: %d", pos))
	}

	l.pos = pos
}

func (l list[T]) first() *T {
	if len(l.data) == 0 {
		return nil
	}

	return &l.data[0]
}

func (l list[T]) last() *T {
	if len(l.data) == 0 {
		return nil
	}

	return &l.data[l.pos-1]
}

// TODO : check bounds in the tests
func (l list[T]) toSlize() []T {
	return l.data[0:l.pos]
}

func remove[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}

	return l
}

// TODO : use bounds in quantifiers?
type bounds struct {
	from, to int
}

func (b bounds) size() int {
	return b.to - b.from
}

type boundsList struct {
	data map[int]Match
	max  Match
	f    func(Match) bounds // TODO : is it really required?
}

func newBoundsList(f func(Match) bounds) *boundsList {
	b := new(boundsList)
	b.data = make(map[int]Match)
	b.f = f
	return b
}

func (b *boundsList) clear() {
	// for key, _ := range b.data {
	// 	delete(b.data, key)
	// }

	b.data = make(map[int]Match, len(b.data))
	b.max = nil
}

func (b *boundsList) push(newMatch Match) {
	x := b.f(newMatch)

	if prevMatch, exists := b.data[x.from]; exists {
		b.data[x.from] = b.longestMatch(prevMatch, newMatch)
	} else {
		b.data[x.from] = newMatch
	}

	if b.max == nil {
		b.max = newMatch
	} else {
		b.max = b.longestMatch(b.max, newMatch)
	}
}

func (b *boundsList) maximum() Match {
	return b.max
}

// TODO : rename to earlist or longest match?
func (b *boundsList) longestMatch(x, y Match) Match {
	xBounds := b.f(x)
	yBounds := b.f(y)

	// TODO : without bounds? only match? without b.f ?

	if xBounds.from < yBounds.from {
		return x
	}

	if xBounds.size() > yBounds.size() {
		return x
	}

	if xBounds.from == yBounds.from {
		return y
	}

	// is it duplication?
	if xBounds.size() > yBounds.size() {
		return x
	}

	return y
}

// todo - better name?
type match struct {
	from int
	to   int
	node node
}

type fullScanner struct {
	onMatch func(node, int, int)
	matches list[match]
	groups      captures
	namedGroups captures
}

func newFullScanner(onMatch func(node, int, int)) *fullScanner {
	s := new(fullScanner)
	s.onMatch = onMatch
	s.matches = *newList[match](100) // pointer?
	s.groups = newCaptures()
	s.namedGroups = newCaptures()
	return s
}

func (s *fullScanner) Match(n node, from, to int, isLeaf, isEmpty bool) {
	m := match{
		from: from,
		to:   to,
		node: n,
	}

	s.matches.append(m)

	if isLeaf {
		s.onMatch(n, from, to)
	}
}

func (s *fullScanner) Position() int {
	return s.matches.size()
}

func (s *fullScanner) Rewind(size int) {
	if s.matches.size() < size {
		return
	}

	s.matches.truncate(size)
}

func (s *fullScanner) FirstMatch() *match {
	if s.matches.size() > 0 {
		return s.matches.first()
	}

	return nil
}

// TODO : add second result like (*match, bool) ?
func (s *fullScanner) LastMatch() *match {
	if s.matches.size() > 0 {
		return s.matches.last()
	}

	return nil
}

func (s *fullScanner) StartNamedGroup(name string, index int) {
	s.namedGroups.From(name, index)
}

// TODO : rename to matchNamedGroup?
func (s *fullScanner) EndNamedGroup(name string, index int) {
	s.namedGroups.To(name, index)
}

func (s *fullScanner) DeleteNamedGroup(name string) {
	s.namedGroups.Delete(name)
}

func (s *fullScanner) StartGroup(name string, index int) {
	s.groups.From(name, index)
}

// TODO : rename to matchGroup?
func (s *fullScanner) EndGroup(name string, index int) {
	s.groups.To(name, index)
}

func (s *fullScanner) DeleteGroup(name string) {
	s.groups.Delete(name)
}

type node interface {
	getKey() string
	getExpressions() dict
	getNestedNodes() index
	isEnd() bool

	match(Handler, int, int, Callback)
	merge(node)
	walk(func(node))
}

type scanor func(node, int, int)

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
			x.getExpressions().add(str)
		}
	})

	key := newNode.getKey()

	if prev, exists := t.nodes[key]; exists {
		prev.merge(newNode)
	} else {
		t.nodes[key] = newNode
	}
}

func (t *trie) Match(text string) ([]FullMatch, error) {
	if len(text) == 0 {
		return nil, nil
	}

	// out := func(match *FullMatch) {
	// 	return nil
	// }

	// cb := func(n node, from int, to int) {

	// }

	// scanner := newScanner(exp, scanner, cb)

	return nil, nil
}

type FullMatch struct {
	data        string
	from        int
	to          int
	groups      []bounds
	namedGroups map[string]bounds
	empty       bool
}

func (t *trie) scan(onMatch scanor) {
	// start := s.matches.first()

	// match := FullMatch{
	// 	// add subString method?
	// 	// data:        s.input,
	// 	from:        start.from,
	// 	to:          to,
	// 	node:        m.node,
	// 	groups:      s.groups.ToSlice(m.to),
	// 	namedGroups: s.groups.ToMap(m.to),
	// 	empty:       isEmpty,
	// }

	// out := func(n node, from, to int) {
	// 	start := 1

	// 	match := FullMatch{
	// 		// add subString method?
	// 		// data:        s.input,
	// 		from:        start.from,
	// 		to:          to,
	// 		node:        m.node,
	// 		groups:      s.groups.ToSlice(m.to),
	// 		namedGroups: s.groups.ToMap(m.to),
	// 		empty:       isEmpty,
	// 	}

	// 	return nil
	// }

	// cb := func(n node, from int, to int) {

	// }

	// scanner := newScanner(exp, scanner)
}

type union struct {
	key       string
	Value     map[string]node   `json:"value,omitempty"`
	lastNodes map[node]struct{} // TODO : interface like key, is it ok?
}

func newUnion(variants []node) *union {
	n := new(union)
	n.Value = make(map[string]node, len(variants))
	n.lastNodes = make(map[node]struct{}, len(variants))

	variantKey := bytes.NewBuffer(nil)
	key := bytes.NewBuffer(nil)

	last := len(variants) - 1

	for i, variant := range variants {
		variant.walk(func(x node) {
			variantKey.WriteString(x.getKey())

			if len(x.getNestedNodes()) == 0 {
				n.lastNodes[x] = struct{}{}
			}
		})

		n.Value[variantKey.String()] = variant
		key.Write(variantKey.Bytes())
		variantKey.Reset()

		if i != last {
			key.WriteRune('|')
		}
	}

	n.key = key.String()

	variantKey.Reset()
	key.Reset()

	return n
}

func (n *union) getKey() string {
	return n.key
}

func (n *union) walk(f func(node)) {
	f(n)

	for _, x := range n.Value {
		x.walk(f)
	}
}

func (n *union) getExpressions() dict {
	for _, x := range n.Value {
		return x.getExpressions()
	}

	return nil
}

func (n *union) getNestedNodes() index {
	return nil
}

func (n *union) isEnd() bool {
	return len(n.getExpressions()) == 0
}

func (n *union) merge(x node) {
	// just ignore it?
	panic(fmt.Sprintf("union can't be merged with : %v", x))
}

func (n *union) match(handler Handler, from, to int, f Callback) {
	// TODO : implement it
	return
}

func (n *union) matchUnion(
	handler Handler,
	id string,
	from, to int,
	onMatch Callback,
	onMatchVariant func(string, int),
) {
	// TODO : check length?

	n.scanVariants(handler, from, to, func(variant node, vFrom, vTo int) {
		onMatchVariant(id, vTo)

		if _, exists := n.lastNodes[variant]; exists {
			handler.Match(variant, vFrom, vTo, n.isEnd(), false) // TODO : what is empty?
			onMatch(variant, vFrom, vTo)

			// TODO : why + 1?
			n.matchNested(handler, vFrom, vTo+1, func(nested node, nFrom, nTo int) {
				onMatch(nested, nFrom, nTo)
				handler.Match(nested, nFrom, nTo, n.isEnd(), false)
			})
		}
	})
}

func (n *union) scanVariants(handler Handler, from, to int, f Callback) {
	for _, variant := range n.Value {
		position := handler.Position()
		variant.match(handler, from, to, f)
		handler.Rewind(position)
	}
}

func (n *union) matchNested(handler Handler, from, to int, f Callback) {
	// how to traverse nested without nested?
	return
}

// is (foo|bar) is equal (bar|foo) ?
// (fo|f)(o|oo)

type group struct {
	uniqID      string
	Value       *union
	Expressions dict  `json:"expression,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *group) getKey() string {
	return fmt.Sprintf("(%s)", n.Value.getKey())
}

func (n *group) getNestedNodes() index {
	return n.Nested
}

func (n *group) getExpressions() dict {
	return n.Expressions
}

func (n *group) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *group) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *group) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *group) match(handler Handler, from, to int, f Callback) {
	handler.StartGroup(n.uniqID, from)
	n.Value.matchUnion(handler, n.uniqID, from, to, f, handler.EndNamedGroup)
	handler.DeleteGroup(n.uniqID)
}

type namedGroup struct {
	Name        string `json:"name,omitempty"`
	Value       *union `json:"value,omitempty"`
	Expressions dict   `json:"expressions,omitempty"`
	Nested      index  `json:"nested,omitempty"`
}

func (n *namedGroup) getKey() string {
	return fmt.Sprintf("(?<%s>%s)", n.Name, n.Value.getKey())
}

func (n *namedGroup) getNestedNodes() index {
	return n.Nested
}

func (n *namedGroup) getExpressions() dict {
	return n.Expressions
}

func (n *namedGroup) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *namedGroup) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *namedGroup) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *namedGroup) match(handler Handler, from, to int, f Callback) {
	handler.StartNamedGroup(n.Name, from)
	n.Value.matchUnion(handler, n.Name, from, to, f, handler.EndNamedGroup)
	handler.DeleteNamedGroup(n.Name)
}

type notCapturedGroup struct {
	Value       *union `json:"value,omitempty"`
	Expressions dict   `json:"expressions,omitempty"`
	Nested      index  `json:"nested,omitempty"`
}

func (n *notCapturedGroup) getKey() string {
	return fmt.Sprintf("(?:%s)", n.Value.getKey())
}

func (n *notCapturedGroup) getNestedNodes() index {
	return n.Nested
}

func (n *notCapturedGroup) getExpressions() dict {
	return n.Expressions
}

func (n *notCapturedGroup) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *notCapturedGroup) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *notCapturedGroup) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *notCapturedGroup) match(handler Handler, from, to int, f Callback) {
	// optimize and remove stub f?
	n.Value.matchUnion(handler, "", from, to, f, func(_ string, _ int) {})
}

type char struct {
	Value       rune  `json:"value,omitempty"`
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *char) getKey() string {
	return string(n.Value)
}

func (n *char) getNestedNodes() index {
	return n.Nested
}

func (n *char) getExpressions() dict {
	return n.Expressions
}

func (n *char) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *char) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *char) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *char) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if x == n.Value {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *char) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

// add something to empty json value, and in another spec symbols
type dot struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *dot) getKey() string {
	return "."
}

func (n *dot) getNestedNodes() index {
	return n.Nested
}

func (n *dot) getExpressions() dict {
	return n.Expressions
}

func (n *dot) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *dot) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *dot) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *dot) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if x != '\n' {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *dot) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type digit struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *digit) getKey() string {
	return "\\d"
}

func (n *digit) getNestedNodes() index {
	return n.Nested
}

func (n *digit) getExpressions() dict {
	return n.Expressions
}

func (n *digit) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *digit) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *digit) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *digit) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *digit) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type nonDigit struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *nonDigit) getKey() string {
	return "\\D"
}

func (n *nonDigit) getNestedNodes() index {
	return n.Nested
}

func (n *nonDigit) getExpressions() dict {
	return n.Expressions
}

func (n *nonDigit) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *nonDigit) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *nonDigit) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *nonDigit) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if !unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *nonDigit) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type word struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *word) getKey() string {
	return "\\w"
}

func (n *word) getNestedNodes() index {
	return n.Nested
}

func (n *word) getExpressions() dict {
	return n.Expressions
}

func (n *word) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *word) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *word) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *word) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *word) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type nonWord struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *nonWord) getKey() string {
	return "\\W"
}

func (n *nonWord) getNestedNodes() index {
	return n.Nested
}

func (n *nonWord) getExpressions() dict {
	return n.Expressions
}

func (n *nonWord) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *nonWord) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *nonWord) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *nonWord) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if !(x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *nonWord) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type space struct {
	Expressions dict  `json:"expression,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *space) getKey() string {
	return "\\s"
}

func (n *space) getNestedNodes() index {
	return n.Nested
}

func (n *space) getExpressions() dict {
	return n.Expressions
}

func (n *space) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *space) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *space) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *space) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if unicode.IsSpace(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *space) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type nonSpace struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *nonSpace) getKey() string {
	return "\\S"
}

func (n *nonSpace) getNestedNodes() index {
	return n.Nested
}

func (n *nonSpace) getExpressions() dict {
	return n.Expressions
}

func (n *nonSpace) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *nonSpace) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *nonSpace) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *nonSpace) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if !unicode.IsSpace(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *nonSpace) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type startOfLine struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *startOfLine) getKey() string {
	return "^"
}

func (n *startOfLine) getNestedNodes() index {
	return n.Nested
}

func (n *startOfLine) getExpressions() dict {
	return n.Expressions
}

func (n *startOfLine) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *startOfLine) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *startOfLine) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *startOfLine) match(handler Handler, from, to int, f Callback) {
	if from == 0 {
		return
	}

	if from == 0 || n.isEndOfLine(handler, from-1) { // TODO : check \n\r too
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *startOfLine) isEndOfLine(handler Handler, idx int) bool {
	x, err := handler.Input().ReadAt(idx)
	if err != nil {
		// TODO : just ignore it?
	}

	return x == '\n'
}

func (n *startOfLine) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type endOfLine struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *endOfLine) getKey() string {
	return "$"
}

func (n *endOfLine) getNestedNodes() index {
	return n.Nested
}

func (n *endOfLine) getExpressions() dict {
	return n.Expressions
}

func (n *endOfLine) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *endOfLine) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *endOfLine) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *endOfLine) match(handler Handler, from, to int, f Callback) {
	// TODO : implement it
	return
}

type startOfString struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *startOfString) getKey() string {
	return "\\A"
}

func (n *startOfString) getNestedNodes() index {
	return n.Nested
}

func (n *startOfString) getExpressions() dict {
	return n.Expressions
}

func (n *startOfString) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *startOfString) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *startOfString) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *startOfString) match(handler Handler, from, to int, f Callback) {
	if from == 0 {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *startOfString) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type endOfString struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *endOfString) getKey() string {
	return "\\z"
}

func (n *endOfString) getNestedNodes() index {
	return n.Nested
}

func (n *endOfString) getExpressions() dict {
	return n.Expressions
}

func (n *endOfString) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *endOfString) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *endOfString) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *endOfString) match(handler Handler, from, to int, f Callback) {
	// TODO : implement it
}

func (n *endOfString) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type rangeNode struct {
	From        rune  `json:"from,omitempty"`
	To          rune  `json:"to,omitempty"`
	Nested      index `json:"nested,omitempty"`
	Expressions dict  `json:"expressions,omitempty"`
}

func (n *rangeNode) getKey() string {
	return string([]rune{n.From, '-', n.To})
}

func (n *rangeNode) getNestedNodes() index {
	return n.Nested
}

func (n *rangeNode) getExpressions() dict {
	return n.Expressions
}

func (n *rangeNode) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *rangeNode) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *rangeNode) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *rangeNode) match(handler Handler, from, to int, f Callback) {
	x, err := handler.Input().ReadAt(from)
	if err != nil {
		// TODO : just ignore it?
		return
	}

	if x >= n.From && x <= n.To {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)
		handler.Rewind(pos)
	}
}

func (n *rangeNode) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type quantifier struct {
	From        int   `json:"from"`
	To          *int  `json:"to,omitempty"`
	More        bool  `json:"more,omitempty"`
	Value       node  `json:"value,omitempty"`
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *quantifier) getKey() string {
	return n.Value.getKey() + n.getQuantifierKey()
}

func (n *quantifier) getQuantifierKey() string {
	if n.From == 0 && n.To == nil && n.More {
		return "*"
	}

	if n.From == 1 && n.To == nil && n.More {
		return "+"
	}

	if n.From == 0 && n.To != nil && *n.To == 1 {
		return "?"
	}

	var b strings.Builder

	b.WriteRune('{')
	b.WriteString(fmt.Sprintf("%d", n.From))

	if n.More {
		b.WriteRune(',')
	}

	if n.To != nil {
		b.WriteString(fmt.Sprintf("%d", *n.To))
	}

	b.WriteRune('}')

	return b.String()
}

func (n *quantifier) getNestedNodes() index {
	return n.Nested
}

func (n *quantifier) getExpressions() dict {
	return n.Expressions
}

func (n *quantifier) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *quantifier) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *quantifier) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *quantifier) match(handler Handler, from, to int, f Callback) {
	start := handler.Position()

	n.recursiveMatch(1, handler, from, to, func(match node, mFrom, mTo int) {
		pos := handler.Position()
		handler.Match(n, from, mTo, n.isEnd(), false)
		f(n, from, mTo)
		n.matchNested(handler, mTo+1, to, f)
		handler.Rewind(pos)
	})

	handler.Rewind(start)

	if n.From == 0 && n.More {
		// TODO : how to return zero match?
		_, mTo := handler.LastMatch()
		ok := true

		if ok {
			handler.Match(n, mTo, mTo, n.isEnd(), false)
		} else {
			handler.Match(n, from, from, n.isEnd(), true)
		}

		n.matchNested(handler, from, to, f)
	}

	handler.Rewind(start)
}

func (n *quantifier) recursiveMatch(count int, handler Handler, from, to int, f Callback) {
	n.Value.match(handler, from, to, func(match node, mFrom, mTo int) {
		if n.To == nil || *n.To >= count {
			if n.inBounds(count) {
				f(match, mFrom, mTo)
			}

			next := count + 1

			if n.To == nil || *n.To >= next {
				n.recursiveMatch(next, handler, mTo+1, to, f)
			}
		}
	})
}

func (n *quantifier) inBounds(q int) bool {
	if n.From > q {
		return false
	}

	if !n.More && n.From != q {
		return false
	}

	if n.To != nil && *n.To > q {
		return false
	}

	return true
}

func (n *quantifier) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type positiveSet struct {
	Value       []node `json:"value,omitempty"`
	Expressions dict   `json:"expressions,omitempty"`
	Nested      index  `json:"nested,omitempty"`
}

func (n *positiveSet) getKey() string {
	subKeys := make([]string, len(n.Value))

	for i, value := range n.Value {
		subKeys[i] = value.getKey()
	}

	sort.Slice(subKeys, func(i, j int) bool {
		return subKeys[i] < subKeys[j]
	})

	x := strings.Join(subKeys, "")

	return fmt.Sprintf("[%s]", x)
}

func (n *positiveSet) getNestedNodes() index {
	return n.Nested
}

func (n *positiveSet) getExpressions() dict {
	return n.Expressions
}

func (n *positiveSet) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *positiveSet) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *positiveSet) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *positiveSet) match(handler Handler, from, to int, f Callback) {
	// TODO : check size
	// TODO : cache isEnd before loop?

	for _, item := range n.Value {
		pos := handler.Position()

		item.match(handler, from, to, func(match node, mFrom, mTo int) {
			handler.Match(n, from, mTo, n.isEnd(), false)
			f(n, from, mTo)
			n.matchNested(handler, mTo+1, to, f)
		})

		handler.Rewind(pos)
	}
}

func (n *positiveSet) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type negativeSet struct {
	Value       []node `json:"value,omitempty"`
	Expressions dict   `json:"expressions,omitempty"`
	Nested      index  `json:"nested,omitempty"`
}

func (n *negativeSet) getKey() string {
	subKeys := make([]string, len(n.Value))

	for i, value := range n.Value {
		subKeys[i] = value.getKey()
	}

	sort.Slice(subKeys, func(i, j int) bool {
		return subKeys[i] < subKeys[j]
	})

	x := strings.Join(subKeys, "")

	return fmt.Sprintf("[^%s]", x)
}

func (n *negativeSet) getNestedNodes() index {
	return n.Nested
}

func (n *negativeSet) getExpressions() dict {
	return n.Expressions
}

func (n *negativeSet) isEnd() bool {
	return len(n.Expressions) == 0
}

func (n *negativeSet) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *negativeSet) merge(other node) {
	n.Nested.merge(other.getNestedNodes())
	n.Expressions.merge(other.getExpressions())
}

func (n *negativeSet) match(handler Handler, from, to int, f Callback) {
	// TODO : check size
	// TODO : cache isEnd before loop?

	for _, item := range n.Value {
		pos := handler.Position()
		matched := false

		item.match(handler, from, to, func(_ node, _, _ int) {
			// TODO : how to propper stop it to avoid pointless iteration?
			matched = true
		})

		if matched {
			handler.Rewind(pos)
			return
		}

		handler.Match(n, from, from, n.isEnd(), false)
		f(n, from, from)
		n.matchNested(handler, from+1, to, f)

		handler.Rewind(pos)
	}
}

func (n *negativeSet) matchNested(handler Handler, from, to int, f Callback) {
	for _, nested := range n.Nested {
		pos := handler.Position()
		nested.match(handler, from, to, f)
		handler.Rewind(pos)
	}
}

type simpleBuffer struct {
	data     []rune
	position int

	// data string
	// positions []int (stack of last positions)
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

func (b *simpleBuffer) ReadAt(idx int) (rune, error) {
	if idx >= len(b.data) {
		return -1, errors.New("out of bounds")
	}

	return b.data[idx], nil
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

// newBuffer - make buffer which can read text on input

var _ c.Buffer[rune, int] = &simpleBuffer{}

func newBuffer(str string) *simpleBuffer {
	b := new(simpleBuffer)
	b.data = []rune(str)
	b.position = 0
	return b
}

type Trie interface {
	Add(...string) error
	Size() int
	MarshalJSON() ([]byte, error)
	String() string
	Match(string) ([]FullMatch, error)
	// IsInclude(string) bool
	// IsMatched(string) (bool, error)
}

var _ Trie = new(trie)

type parser = c.Combinator[rune, int, node]

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
	// TODO : is it pointless condition?
	// check it on becnmarks / tests
	if _, exists := d[str]; !exists {
		d[str] = struct{}{}
	}
}

func (d dict) merge(other dict) {
	for key, value := range other {
		if _, exists := d[key]; !exists {
			d[key] = value
		}
	}
}

var (
	defaultParser = parseRegexp()
	none          = struct{}{}

	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

func Trace[T any, P any, S any](
	m string,
	parse c.Combinator[T, P, S],
) c.Combinator[T, P, S] {
	return func(buffer c.Buffer[T, P]) (S, error) {
		fmt.Printf("%v\n", m)
		fmt.Printf("%s %v\n", m, buffer)
		fmt.Printf("\t%s position before: %v\n", m, buffer.Position())

		result, err := parse(buffer)
		fmt.Printf("\t%s position after: %v\n", m, buffer.Position())
		if err != nil {
			fmt.Printf("\t%s not parsed: %v %v\n", m, result, err)
			return *new(S), err
		}

		fmt.Println("\tparsed:", fmt.Sprintf("%#v", result))
		return result, err
	}
}

func SkipString(data string) c.Combinator[rune, int, struct{}] {
	return func(buf c.Buffer[rune, int]) (struct{}, error) {
		l := len(data)
		for _, x := range data {
			r, err := buf.Read(true)
			if err != nil {
				return none, err
			}
			if x != r {
				return none, c.NothingMatched
			}
			l -= 1
		}

		if l != 0 {
			return none, c.NothingMatched
		}

		return none, nil
	}
}

// TODO : return error for invalid escaped chars like '\x' (check on rubular)

func parseRegexp() parser {
	var parseExpression parser
	var parseNestedExpression parser

	sep := c.Eq[rune, int]('|')

	// parse union
	union := func(buf c.Buffer[rune, int]) (*union, error) {
		variant, err := parseNestedExpression(buf)
		if err != nil {
			return nil, err
		}

		variants := make([]node, 0, 1)
		variants = append(variants, variant)

		for !buf.IsEOF() {
			pos := buf.Position()

			_, err = sep(buf)
			if err != nil {
				buf.Seek(pos)
				break // return error instead break?
			}

			variant, err = parseNestedExpression(buf)
			if err != nil {
				buf.Seek(pos)
				break // return error instead break?
			}

			variants = append(variants, variant)
		}

		// TODO : check length and eof

		return newUnion(variants), nil
	}

	// TODO : return error for invalid escaped chars like '\x' (check on rubular)

	// parse node
	parseNode := parseOptionalQuantifier(
		choice(
			parseSet('|'),
			parseNotCapturedGroup(union),
			parseNamedGroup(union),
			parseGroup(union),
			parseInvalidQuantifier(),
			parseEscapedMetaCharacters(),
			parseMetaCharacters(),
			parseEscapedSpecSymbols(),
			parseCharacter('|'),
		),
	)

	// parse node of nested expression
	parseNestedNode := parseOptionalQuantifier(
		choice(
			parseSet('|', ')'),
			parseNotCapturedGroup(union),
			parseNamedGroup(union),
			parseGroup(union),
			parseInvalidQuantifier(),
			parseEscapedMetaCharacters(),
			parseMetaCharacters(),
			parseEscapedSpecSymbols(),
			parseCharacter('|', ')'),
		),
	)

	parseExpression = func(buf c.Buffer[rune, int]) (node, error) {
		first, err := parseNode(buf)
		if err != nil {
			return nil, err
		}

		last := first

		for !buf.IsEOF() {
			pos := buf.Position()

			next, err := parseNode(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			last.getNestedNodes()[next.getKey()] = next
			last = next
		}

		return first, nil
	}

	parseNestedExpression = func(buf c.Buffer[rune, int]) (node, error) {
		first, err := parseNestedNode(buf)
		if err != nil {
			return nil, err
		}

		last := first

		for !buf.IsEOF() {
			pos := buf.Position()

			next, err := parseNestedNode(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			last.getNestedNodes()[next.getKey()] = next
			last = next
		}

		return first, nil
	}

	// parse union or expression
	return func(buf c.Buffer[rune, int]) (node, error) {
		expression, err := parseExpression(buf)
		if err != nil {
			return nil, err
		}
		if buf.IsEOF() {
			return expression, nil
		}

		variants := make([]node, 0, 1)
		variants = append(variants, expression)

		for !buf.IsEOF() {
			_, err = sep(buf)
			if err != nil {
				return nil, err
			}

			expression, err = parseExpression(buf)
			if err != nil {
				return nil, err
			}

			variants = append(variants, expression)
		}

		return newUnion(variants), nil
	}
}

func parseSet(except ...rune) parser {
	// TODO : without except?
	parseNode := choice(
		parseRange(append(except, ']')...),
		parseEscapedMetaCharacters(),
		parseEscapedSpecSymbols(),
		parseCharacter(append(except, ']')...),
	)

	return choice(
		parseNegativeSet(parseNode),
		parsePositiveSet(parseNode),
	)
}

func choice(parsers ...parser) parser {
	attempts := make([]parser, len(parsers))

	for i, parse := range parsers {
		attempts[i] = c.Try(parse)
	}

	return c.Choice(attempts...)
}

func between[T any, S any](
	before c.Combinator[rune, int, S],
	body c.Combinator[rune, int, T],
	after c.Combinator[rune, int, S],
) c.Combinator[rune, int, T] {
	return c.Between(before, body, after)
}

func parens[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return between(
		c.Eq[rune, int]('('),
		body,
		c.Eq[rune, int](')'),
	)
}

func angles[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return between(
		c.Eq[rune, int]('<'),
		body,
		c.Eq[rune, int]('>'),
	)
}

func squares[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return between(
		c.Eq[rune, int]('['),
		body,
		c.Eq[rune, int](']'),
	)
}

func number() c.Combinator[rune, int, int] {
	digit := c.Try[rune, int](c.Range[rune, int]('0', '9'))
	zero := rune('0')

	return func(buf c.Buffer[rune, int]) (int, error) {
		token, err := digit(buf)
		if err != nil {
			return 0, err
		}

		result := int(token - zero)
		for {
			token, err = digit(buf)
			if err != nil {
				break
			}

			result = result * 10
			result += int(token - zero)
		}

		return result, nil
	}
}

func parseEscapedSpecSymbols() parser {
	symbols := ".?+*^$[]{}()"
	cases := make(map[rune]parser)

	for _, r := range symbols {
		cases[r] = func(buf c.Buffer[rune, int]) (node, error) {
			x := char{
				Value:       r,
				Nested:      make(index, 0),
				Expressions: make(dict, 0),
			}

			return &x, nil
		}
	}

	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			cases,
			c.Any[rune, int](),
		),
	)
}

func parseInvalidQuantifier() parser {
	invalidChars := map[rune]struct{}{
		'?': {},
		'*': {},
		'+': {},
	}

	return func(buf c.Buffer[rune, int]) (node, error) {
		x, err := buf.Read(false)
		if err != nil {
			return nil, err
		}

		if _, exists := invalidChars[x]; exists {
			return nil, InvalidQuantifierError
		}

		return nil, c.NothingMatched
	}
}

func parseOptionalQuantifier(expression parser) parser {
	any := c.Any[rune, int]()
	digit := c.Try(number())
	comma := c.Try(c.Eq[rune, int](','))
	rightBrace := c.Eq[rune, int]('}')

	parseQuantifier := c.Try(
		c.MapAs(
			map[rune]c.Combinator[rune, int, quantifier]{
				'?': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					q.From = 0
					to := 1
					q.To = &to
					q.More = false

					return q, nil
				},
				'+': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					q.From = 1
					q.More = true

					return q, nil
				},
				'*': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					q.From = 0
					q.More = true

					return q, nil
				},
				'{': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					from, err := digit(buf)
					if err != nil {
						return q, err
					}
					q.From = from

					_, err = comma(buf)
					if err != nil {
						_, err = rightBrace(buf)
						if err != nil {
							return q, err
						}

						return q, nil
					}
					q.More = true

					to, err := digit(buf)
					if err != nil {
						_, err = rightBrace(buf)
						if err != nil {
							return q, err
						}

						return q, err
					}

					// TODO : check it
					if from == 0 && to == 0 {
						return q, c.NothingMatched
					}

					q.To = &to

					// compact {1,1} to {1}
					if from == to {
						q.More = false
						q.To = nil
					}

					_, err = rightBrace(buf)
					if err != nil {
						return q, err
					}

					q.Expressions = make(dict, 0)

					return q, nil
				},
			},
			any,
		),
	)

	return func(buf c.Buffer[rune, int]) (node, error) {
		x, err := expression(buf)
		if err != nil {
			return nil, err
		}

		q, err := parseQuantifier(buf)
		if err != nil {
			return x, nil
		}

		q.Value = x
		q.Nested = make(index, 0)
		q.Expressions = make(dict, 0)

		return &q, nil
	}
}

func parseCharacter(except ...rune) parser {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node, error) {
		c, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := char{
			Value:       c,
			Expressions: make(dict, 0),
			Nested:      make(index, 0),
		}

		return &x, nil
	}
}

func parseMetaCharacters() parser {
	return c.MapAs(
		map[rune]c.Combinator[rune, int, node]{
			'.': func(buf c.Buffer[rune, int]) (node, error) {
				x := dot{
					Expressions: make(dict, 0),
					Nested:      make(index, 0),
				}

				return &x, nil
			},
			'^': func(buf c.Buffer[rune, int]) (node, error) {
				x := startOfLine{
					Expressions: make(dict, 0),
					Nested:      make(index, 0),
				}

				return &x, nil
			},
			'$': func(buf c.Buffer[rune, int]) (node, error) {
				x := endOfLine{
					Expressions: make(dict, 0),
					Nested:      make(index, 0),
				}

				return &x, nil
			},
		},
		c.Any[rune, int](),
	)
}

func parseEscapedMetaCharacters() parser {
	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			map[rune]c.Combinator[rune, int, node]{
				'd': func(buf c.Buffer[rune, int]) (node, error) {
					x := digit{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
				'D': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonDigit{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
				'w': func(buf c.Buffer[rune, int]) (node, error) {
					x := word{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
				'W': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonWord{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
				's': func(buf c.Buffer[rune, int]) (node, error) {
					x := space{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
				'S': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonSpace{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
				'A': func(buf c.Buffer[rune, int]) (node, error) {
					x := startOfString{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
				'z': func(buf c.Buffer[rune, int]) (node, error) {
					x := endOfString{
						Nested:      make(index, 0),
						Expressions: make(dict, 0),
					}

					return &x, nil
				},
			},
			c.Any[rune, int](),
		),
	)
}

func parseGroup(parse c.Combinator[rune, int, *union]) parser {
	return parens(
		func(buf c.Buffer[rune, int]) (node, error) {
			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			x := &group{
				Nested: make(index, 0),
			}

			// TODO : is it good enough for ID?
			x.uniqID = fmt.Sprintf("%p", x)
			x.Value = value
			x.Expressions = make(dict, 0)

			return x, nil
		},
	)
}

func parseNotCapturedGroup(parse c.Combinator[rune, int, *union]) parser {
	before := SkipString("?:")

	return parens(
		func(buf c.Buffer[rune, int]) (node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			x := notCapturedGroup{
				Value:       value,
				Nested:      make(index, 0),
				Expressions: make(dict, 0),
			}

			return &x, nil
		},
	)
}

func parseNamedGroup(parse c.Combinator[rune, int, *union], except ...rune) parser {
	groupName := c.Skip(
		c.Eq[rune, int]('?'),
		angles(
			c.Some(
				0,
				c.Try(c.NoneOf[rune, int](append(except, '>')...)),
			),
		),
	)

	return parens(
		func(buf c.Buffer[rune, int]) (node, error) {
			name, err := groupName(buf)
			if err != nil {
				return nil, err
			}

			variants, err := parse(buf)
			if err != nil {
				return nil, err
			}

			x := namedGroup{
				Name:        string(name),
				Value:       variants,
				Nested:      make(index, 0),
				Expressions: make(dict, 0),
			}

			return &x, nil
		},
	)
}

func parseNegativeSet(expression parser) parser {
	parse := squares(
		c.Skip(
			c.Eq[rune, int]('^'),
			c.Some(1, expression),
		),
	)

	return func(buf c.Buffer[rune, int]) (node, error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := negativeSet{
			Value:       set,
			Nested:      make(index, 0),
			Expressions: make(dict, 0),
		}

		return &x, nil
	}
}

func parsePositiveSet(expression parser) parser {
	parse := squares(c.Some(1, expression))

	return func(buf c.Buffer[rune, int]) (node, error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := positiveSet{
			Value:       set,
			Nested:      make(index, 0),
			Expressions: make(dict, 0),
		}

		return &x, nil
	}
}

func parseRange(except ...rune) parser {
	item := c.NoneOf[rune, int](except...)
	sep := c.Eq[rune, int]('-')

	return func(buf c.Buffer[rune, int]) (node, error) {
		f, err := item(buf)
		if err != nil {
			return nil, err
		}

		_, err = sep(buf)
		if err != nil {
			return nil, err
		}

		t, err := item(buf)
		if err != nil {
			return nil, err
		}

		x := rangeNode{
			From:   f,
			To:     t,
			Nested: make(index, 0),
		}

		return &x, nil
	}
}
