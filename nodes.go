package regular

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type node interface {
	getKey() string
	getExpressions() dict
	addExpression(string)
	getNestedNodes() index
	isEnd() bool

	match(Handler, TextBuffer, int, int, Callback)
	merge(node)
	walk(func(node))
}

type Callback func(x node, from int, to int, empty bool)

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

type nestedNode struct {
	Expressions dict  `json:"expressions,omitempty"`
	Nested      index `json:"nested,omitempty"`
}

func (n *nestedNode) getNestedNodes() index {
	return n.Nested
}

func (n *nestedNode) getExpressions() dict {
	return n.Expressions
}

func (n *nestedNode) addExpression(exp string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	n.Expressions.add(exp)
}

func (n *nestedNode) isEnd() bool {
	return len(n.Expressions) > 0
}

func (n *nestedNode) merge(other node) {
	n.Nested.merge(other.getNestedNodes())

	if n.Expressions == nil {
		n.Expressions = other.getExpressions()
	} else {
		n.Expressions.merge(other.getExpressions())
	}
}

func (n *nestedNode) matchNested(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	pos := handler.Position()

	for _, nested := range n.Nested {
		nested.match(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
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

func (n *union) addExpression(exp string) {
	for _, x := range n.Value {
		x.addExpression(exp)
	}
}

func (n *union) getNestedNodes() index {
	return nil
}

func (n *union) isEnd() bool {
	return len(n.getExpressions()) == 0
}

func (n *union) merge(x node) {
	panic(fmt.Sprintf("union can't be merged with : %v", x))
}

func (n *union) match(_ Handler, _ TextBuffer, _, _ int, _ Callback) {
	panic("not implemented")
}

func (n *union) matchUnion(
	handler Handler,
	input TextBuffer,
	from, to int,
	onMatch Callback,
) {
	n.scanVariants(handler, input, from, to, func(variant node, vFrom, vTo int, empty bool) {
		if _, exists := n.lastNodes[variant]; exists {
			onMatch(variant, vFrom, vTo, empty)
		}
	})
}

func (n *union) scanVariants(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	position := handler.Position()

	for _, variant := range n.Value {
		variant.match(handler, input, from, to, onMatch)
		handler.Rewind(position)
	}
}

// is (foo|bar) is equal (bar|foo) ?
// (fo|f)(o|oo)

type group struct {
	// TODO : it's not really uniq id
	// because the same union in another group is possible
	// probable use node interface like key for map
	uniqID string
	Value  *union `json:"value,omitempty"`
	*nestedNode
}

func (n *group) getKey() string {
	return fmt.Sprintf("(%s)", n.Value.getKey())
}

func (n *group) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *group) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	handler.AddGroup(n.uniqID, from)
	n.Value.matchUnion(
		handler,
		input,
		from,
		to,
		func(_ node, vFrom, vTo int, empty bool) {
			handler.MatchGroup(n.uniqID, vTo)
			handler.Match(n, from, vTo, n.isEnd(), false)
			onMatch(n, from, vTo, empty)
			n.matchNested(handler, input, vTo+1, to, onMatch)
		},
	)
	handler.DeleteGroup(n.uniqID)
}

type namedGroup struct {
	Name  string `json:"name,omitempty"`
	Value *union `json:"value,omitempty"`
	*nestedNode
}

func (n *namedGroup) getKey() string {
	return fmt.Sprintf("(?<%s>%s)", n.Name, n.Value.getKey())
}

func (n *namedGroup) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *namedGroup) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	handler.AddNamedGroup(n.Name, from)
	n.Value.matchUnion(
		handler,
		input,
		from,
		to,
		func(_ node, vFrom, vTo int, empty bool) {
			handler.MatchNamedGroup(n.Name, vTo)
			handler.Match(n, from, vTo, n.isEnd(), false)
			onMatch(n, from, vTo, empty)
			n.matchNested(handler, input, vTo+1, to, onMatch)
		},
	)
	handler.DeleteNamedGroup(n.Name)
}

type notCapturedGroup struct {
	Value *union `json:"value,omitempty"`
	*nestedNode
}

func (n *notCapturedGroup) getKey() string {
	return fmt.Sprintf("(?:%s)", n.Value.getKey())
}

func (n *notCapturedGroup) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *notCapturedGroup) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	n.Value.matchUnion(
		handler,
		input,
		from,
		to,
		func(_ node, vFrom, vTo int, empty bool) {
			handler.Match(n, from, vTo, n.isEnd(), false)
			onMatch(n, from, vTo, empty)
			n.matchNested(handler, input, vTo+1, to, onMatch)
		},
	)
}

type char struct {
	Value rune `json:"value,omitempty"`
	*nestedNode
}

func (n *char) getKey() string {
	return string(n.Value)
}

func (n *char) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *char) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) == n.Value {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

// add something to empty json value, and in another spec symbols
type dot struct {
	*nestedNode
}

func (n *dot) getKey() string {
	return "."
}

func (n *dot) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *dot) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) != '\n' {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type digit struct {
	*nestedNode
}

func (n *digit) getKey() string {
	return "\\d"
}

func (n *digit) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *digit) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type nonDigit struct {
	*nestedNode
}

func (n *nonDigit) getKey() string {
	return "\\D"
}

func (n *nonDigit) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *nonDigit) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type word struct {
	*nestedNode
}

func (n *word) getKey() string {
	return "\\w"
}

func (n *word) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *word) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type nonWord struct {
	*nestedNode
}

func (n *nonWord) getKey() string {
	return "\\W"
}

func (n *nonWord) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *nonWord) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !(x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type space struct {
	*nestedNode
}

func (n *space) getKey() string {
	return "\\s"
}

func (n *space) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *space) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsSpace(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type nonSpace struct {
	*nestedNode
}

func (n *nonSpace) getKey() string {
	return "\\S"
}

func (n *nonSpace) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *nonSpace) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsSpace(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type startOfLine struct {
	*nestedNode
}

func (n *startOfLine) getKey() string {
	return "^"
}

func (n *startOfLine) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *startOfLine) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	// TODO : precache new line positions in buffer?

	if from == 0 || n.isEndOfLine(input, from-1) { // TODO : check \n\r too
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.matchNested(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
}

func (n *startOfLine) isEndOfLine(input TextBuffer, idx int) bool {
	if idx < 0 {
		return false
	}

	x := input.ReadAt(idx)

	switch x {
	case '\n':
		return true
	case '\r':
		if idx == 0 {
			return true
		}

		// TODO : looks strange
		return input.ReadAt(idx-1) == '\n'
	default:
		return false
	}
}

type endOfLine struct {
	*nestedNode
}

func (n *endOfLine) getKey() string {
	return "$"
}

func (n *endOfLine) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *endOfLine) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	// TODO : precache new line positions in buffer?

	if n.isEndOfLine(input, from) { // TODO : check \n\r too
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.matchNested(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
}

func (n *endOfLine) isEndOfLine(input TextBuffer, idx int) bool {
	if idx > input.Size() {
		return false
	}

	if idx == input.Size() {
		return true
	}

	return input.ReadAt(idx) == '\n'
}

type startOfString struct {
	*nestedNode
}

func (n *startOfString) getKey() string {
	return "\\A"
}

func (n *startOfString) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *startOfString) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from == 0 {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.matchNested(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
}

type endOfString struct {
	*nestedNode
}

func (n *endOfString) getKey() string {
	return "\\z"
}

func (n *endOfString) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *endOfString) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from == input.Size() {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.matchNested(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
}

type rangeNode struct {
	From rune `json:"from,omitempty"`
	To   rune `json:"to,omitempty"`
	*nestedNode
}

func (n *rangeNode) getKey() string {
	return string([]rune{n.From, '-', n.To})
}

func (n *rangeNode) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *rangeNode) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if x >= n.From && x <= n.To {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)
		handler.Rewind(pos)
	}
}

type quantifier struct {
	From  int  `json:"from"`
	To    *int `json:"to,omitempty"`
	More  bool `json:"more,omitempty"`
	Value node `json:"value,omitempty"`
	*nestedNode
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

func (n *quantifier) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *quantifier) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	start := handler.Position()

	n.recursiveMatch(1, handler, input, from, to, func(match node, mFrom, mTo int, empty bool) {
		pos := handler.Position()
		handler.Match(n, from, mTo, n.isEnd(), false)
		onMatch(n, from, mTo, empty)
		n.matchNested(handler, input, mTo+1, to, onMatch)
		handler.Rewind(pos)
	})

	handler.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.From == 0 {
		m := handler.LastMatch()

		if m != nil {
			// TODO : remove condition and this line?
			handler.Match(n, m.to, m.to, n.isEnd(), false)
		} else {
			handler.Match(n, from, from, n.isEnd(), true)
		}

		n.matchNested(handler, input, from, to, onMatch)
	}

	handler.Rewind(start)
}

func (n *quantifier) recursiveMatch(
	count int,
	handler Handler,
	input TextBuffer,
	from, to int,
	onMatch Callback,
) {
	n.Value.match(handler, input, from, to, func(match node, mFrom, mTo int, empty bool) {
		if n.To == nil || *n.To >= count {
			if n.inBounds(count) {
				onMatch(match, mFrom, mTo, empty)
			}

			next := count + 1

			if n.To == nil || *n.To >= next {
				n.recursiveMatch(next, handler, input, mTo+1, to, onMatch)
			}
		}
	})
}

func (n *quantifier) inBounds(q int) bool {
	if n.From > q {
		return false
	}

	if n.More {
		return true
	}

	if n.To != nil {
		return q <= *n.To
	}

	return n.From == q
}

type characterClass struct {
	Value []node `json:"value,omitempty"`
	*nestedNode
}

func (n *characterClass) getKey() string {
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

func (n *characterClass) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *characterClass) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	// TODO : cache isEnd before loop?

	pos := handler.Position()

	for _, item := range n.Value {
		item.match(handler, input, from, to, func(match node, mFrom, mTo int, empty bool) {
			handler.Match(n, from, mTo, n.isEnd(), false)
			onMatch(n, from, mTo, empty)
			n.matchNested(handler, input, mTo+1, to, onMatch)
		})

		handler.Rewind(pos)
	}
}

type negatedCharacterClass struct {
	Value []node `json:"value,omitempty"`
	*nestedNode
}

func (n *negatedCharacterClass) getKey() string {
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

func (n *negatedCharacterClass) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *negatedCharacterClass) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	// TODO : cache isEnd before loop?

	pos := handler.Position()

	for _, item := range n.Value {
		matched := false

		item.match(handler, input, from, to, func(_ node, _, _ int, _ bool) {
			// TODO : how to propper stop it to avoid pointless iteration?
			matched = true
		})

		if matched {
			handler.Rewind(pos)
			return
		}

		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.matchNested(handler, input, from+1, to, onMatch)

		handler.Rewind(pos)
	}
}
