package regular

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type node interface {
	getKey() string
	getExpressions() dict
	addExpression(string)
	getNestedNodes() index
	isEnd() bool

	scan(Handler, TextBuffer, int, int, Callback)
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

func newDict(items ...string) dict {
	d := make(dict)
	for _, x := range items {
		d.add(x)
	}
	return d
}

func (d dict) add(str string) {
	d[str] = struct{}{}
}

func (d dict) merge(other dict) dict {
	for key, value := range other {
		d[key] = value
	}

	return d
}

func (d dict) Slice() []string {
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

func (n *nestedNode) match(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	pos := handler.Position()

	for _, nested := range n.Nested {
		nested.scan(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
}

// https://www.regular-expressions.info/posix.html
//
// - what is better behaviour, first match or longest match?
// - it's important for compaction

// https://www.regular-expressions.info/alternation.html
//
// Remember That The Regex Engine Is Eager
//
// The consequence is that in certain situations, the order of the alternatives matters.
// With expression "Get|GetValue|Set|SetValue" and string SetValue,
// should be matched third variant - "Set"
//
// TODO : add test for if it possible

// BUT

// POSIX ERE Alternation Returns The Longest Match

// In the tutorial topic about alternation, I explained that the regex engine will stop as soon as it finds a matching alternative.
// The POSIX standard, however, mandates that the longest match be returned.
// When applying Set|SetValue to SetValue, a POSIX-compliant regex engine will match SetValue entirely.
// Even if the engine is a regex-directed NFA engine, POSIX requires that it simulates DFA text-directed matching by trying all alternatives,
// and returning the longest match, in this case SetValue.
// A traditional NFA engine would match Set, as do all other regex flavors discussed on this website.

// A POSIX-compliant engine will still find the leftmost match.
// If you apply Set|SetValue to Set or SetValue once, it will match Set.
// The first position in the string is the leftmost position where our regex can find a valid match.
// The fact that a longer match can be found further in the string is irrelevant.
// If you apply the regex a second time, continuing at the first space in the string, then SetValue will be matched.
// A traditional NFA engine would match Set at the start of the string as the first match, and Set at the start of the 3rd word in the string as the second match.

type alternation struct {
	Value     map[string]node   `json:"value,omitempty"`
	lastNodes map[node]struct{} // TODO : interface like key, is it ok?
	*nestedNode
}

func newAlternation(variants []node) *alternation {
	n := new(alternation)
	n.Value = make(map[string]node, len(variants))
	n.lastNodes = make(map[node]struct{}, len(variants))
	n.nestedNode = newNestedNode()

	variantKey := bytes.NewBuffer(nil)

	for _, variant := range variants {
		variant.walk(func(x node) {
			variantKey.WriteString(x.getKey())

			if len(x.getNestedNodes()) == 0 {
				n.lastNodes[x] = struct{}{}
			}
		})

		x := variantKey.String()
		n.Value[x] = variant
		variantKey.Reset()
	}

	variantKey.Reset()

	return n
}

func (n *alternation) getKey() string {
	variantKeys := make([]string, 0, len(n.Value))

	for _, variant := range n.Value {
		variantKeys = append(variantKeys, variant.getKey())
	}

	return strings.Join(variantKeys, ",")
}

func (n *alternation) walk(f func(node)) {
	f(n)

	for _, x := range n.Value {
		x.walk(f)
	}
}

// TODO : check it without groups too
func (n *alternation) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	n.scanVariants(
		handler,
		input,
		from,
		to,
		func(_ node, vFrom, vTo int, empty bool) {
			handler.Match(n, from, vTo, n.isEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.match(handler, input, vTo+1, to, onMatch)
		},
	)
}

func (n *alternation) scanAlternation(
	handler Handler,
	input TextBuffer,
	from, to int,
	onMatch Callback,
) {
	n.scanVariants(
		handler,
		input,
		from,
		to,
		func(variant node, vFrom, vTo int, empty bool) {
			if _, exists := n.lastNodes[variant]; exists {
				onMatch(variant, vFrom, vTo, empty)
			}
		},
	)
}

func (n *alternation) scanVariants(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	position := handler.Position()

	for _, variant := range n.Value {
		variant.scan(handler, input, from, to, onMatch)
		handler.Rewind(position)
	}
}

type group struct {
	// TODO : it's not really uniq id
	uniqID string
	Value  *alternation `json:"value,omitempty"`
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

func (n *group) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	handler.AddGroup(n.uniqID, from)
	n.Value.scanAlternation(
		handler,
		input,
		from,
		to,
		func(_ node, vFrom, vTo int, empty bool) {
			handler.MatchGroup(n.uniqID, vTo)
			// a lot of line like belowe, maybe move it in handler or trie?
			handler.Match(n, from, vTo, n.isEnd(), false) // is it possible to remove and use only onMatch?
			onMatch(n, from, vTo, empty)
			n.nestedNode.match(handler, input, vTo+1, to, onMatch)
		},
	)
	handler.DeleteGroup(n.uniqID)
}

type namedGroup struct {
	Name  string       `json:"name,omitempty"`
	Value *alternation `json:"value,omitempty"`
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

func (n *namedGroup) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	handler.AddNamedGroup(n.Name, from)
	n.Value.scanAlternation(
		handler,
		input,
		from,
		to,
		func(_ node, vFrom, vTo int, empty bool) {
			handler.MatchNamedGroup(n.Name, vTo)
			handler.Match(n, from, vTo, n.isEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.match(handler, input, vTo+1, to, onMatch)
		},
	)
	handler.DeleteNamedGroup(n.Name)
}

type notCapturedGroup struct {
	Value *alternation `json:"value,omitempty"`
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

func (n *notCapturedGroup) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	n.Value.scanAlternation(
		handler,
		input,
		from,
		to,
		func(_ node, vFrom, vTo int, empty bool) {
			handler.Match(n, from, vTo, n.isEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.match(handler, input, vTo+1, to, onMatch)
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

func (n *char) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) == n.Value {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *dot) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) != '\n' {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *digit) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *nonDigit) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *word) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *nonWord) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !(x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *space) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsSpace(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *nonSpace) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsSpace(x) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)
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

func (n *startOfLine) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	// TODO : precache new line positions in buffer?

	if from == 0 || n.isEndOfLine(input, from-1) { // TODO : check \n\r too
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.match(handler, input, from, to, onMatch)
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

func (n *endOfLine) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	// TODO : precache new line positions in buffer?

	if n.isEndOfLine(input, from) {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.match(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
}

// TODO : check \n\r too
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

func (n *startOfString) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from == 0 {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.match(handler, input, from, to, onMatch)
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

func (n *endOfString) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from == input.Size() {
		pos := handler.Position()
		handler.Match(n, from, from, n.isEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.match(handler, input, from, to, onMatch)
		handler.Rewind(pos)
	}
}

// https://www.regular-expressions.info/repeat.html

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
	} else if n.To != nil {
		b.WriteRune(',')
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

func (n *quantifier) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	start := handler.Position()

	n.recursiveScan(1, handler, input, from, to, func(_ node, _, mTo int, empty bool) {
		pos := handler.Position()
		handler.Match(n, from, mTo, n.isEnd(), false)
		onMatch(n, from, mTo, empty)
		n.nestedNode.match(handler, input, mTo+1, to, onMatch)
		handler.Rewind(pos)
	})

	handler.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.From == 0 {
		m := handler.LastMatch()

		if m != nil {
			// TODO : remove condition and this line?
			handler.Match(n, m.span.to, m.span.to, n.isEnd(), false)
		} else {
			handler.Match(n, from, from, n.isEnd(), true)
		}

		n.nestedNode.match(handler, input, from, to, onMatch)
	}

	handler.Rewind(start)
}

func (n *quantifier) recursiveScan(
	count int,
	handler Handler,
	input TextBuffer,
	from, to int,
	onMatch Callback,
) {
	n.Value.scan(handler, input, from, to, func(match node, mFrom, mTo int, empty bool) {
		if n.To == nil || *n.To >= count {
			if n.inBounds(count) {
				onMatch(match, mFrom, mTo, empty)
			}

			next := count + 1

			if n.To == nil || *n.To >= next {
				n.recursiveScan(next, handler, input, mTo+1, to, onMatch)
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

// https://www.regular-expressions.info/charclass.html

type characterClass struct {
	table *unicode.RangeTable
	*nestedNode
}

func (n *characterClass) getKey() string {
	b := new(strings.Builder)

	b.WriteString("Class[R16(")

	for _, r := range n.table.R16 {
		b.WriteString(fmt.Sprintf("%d", r.Lo))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Hi))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Stride))
		b.WriteString(",")
	}

	b.WriteString("),R32(")

	for _, r := range n.table.R32 {
		b.WriteString(fmt.Sprintf("%d", r.Lo))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Hi))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Stride))
		b.WriteString(",")
	}

	b.WriteString(")]")

	return b.String()
}

func (n *characterClass) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *characterClass) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	// TODO : always only one character?
	if unicode.In(x, n.table) {
		pos := handler.Position()

		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)

		handler.Rewind(pos)
	}
}

type negatedCharacterClass struct {
	table *unicode.RangeTable
	*nestedNode
}

func (n *negatedCharacterClass) getKey() string {
	b := new(strings.Builder)

	b.WriteString("NegatedClass[R16(")

	for _, r := range n.table.R16 {
		b.WriteString(fmt.Sprintf("%d", r.Lo))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Hi))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Stride))
		b.WriteString(",")
	}

	b.WriteString("),R32(")

	for _, r := range n.table.R32 {
		b.WriteString(fmt.Sprintf("%d", r.Lo))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Hi))
		b.WriteString("-")
		b.WriteString(fmt.Sprintf("%d", r.Stride))
		b.WriteString(",")
	}

	b.WriteString(")]")

	return b.String()
}

func (n *negatedCharacterClass) walk(f func(node)) {
	f(n)

	for _, x := range n.Nested {
		x.walk(f)
	}
}

func (n *negatedCharacterClass) scan(handler Handler, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	// TODO : always only one character?
	if !unicode.In(x, n.table) {
		pos := handler.Position()

		handler.Match(n, from, from, n.isEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.match(handler, input, from+1, to, onMatch)

		handler.Rewind(pos)
	}
}
