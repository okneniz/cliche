package cliche

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Node interface {
	GetKey() string
	GetExpressions() Set
	AddExpression(string)
	GetNestedNodes() map[string]Node
	IsEnd() bool

	Scan(Output, TextBuffer, int, int, Callback)
	Merge(Node)
	Traverse(func(Node))
}

type Callback func(x Node, from int, to int, empty bool)

// TODO : rename to Set
type Set map[string]struct{}

func newSet(items ...string) Set {
	d := make(Set)
	for _, x := range items {
		d.add(x)
	}
	return d
}

func (d Set) add(str string) {
	d[str] = struct{}{}
}

func (d Set) merge(other Set) Set {
	for key, value := range other {
		d[key] = value
	}

	return d
}

func (d Set) Slice() []string {
	result := make([]string, len(d))
	i := 0
	for key := range d {
		result[i] = key
		i++
	}
	return result
}

type nestedNode struct {
	Expressions Set             `json:"expressions,omitempty"`
	Nested      map[string]Node `json:"nested,omitempty"`
}

func (n *nestedNode) GetNestedNodes() map[string]Node {
	return n.Nested
}

func (n *nestedNode) GetExpressions() Set {
	return n.Expressions
}

func (n *nestedNode) AddExpression(exp string) {
	if n.Expressions == nil {
		n.Expressions = make(Set)
	}

	n.Expressions.add(exp)
}

func (n *nestedNode) IsEnd() bool {
	return len(n.Expressions) > 0
}

func (n *nestedNode) Merge(other Node) {
	for key, newNode := range other.GetNestedNodes() {
		if prev, exists := n.Nested[key]; exists {
			prev.Merge(newNode)
		} else {
			n.Nested[key] = newNode
		}
	}

	if n.Expressions == nil {
		n.Expressions = other.GetExpressions()
	} else {
		n.Expressions.merge(other.GetExpressions())
	}
}

func (n *nestedNode) Match(
	output Output,
	input TextBuffer,
	from, to int,
	onMatch Callback,
) {
	pos := output.Position()

	for _, nested := range n.Nested {
		nested.Scan(output, input, from, to, onMatch)
		output.Rewind(pos)
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
	Value     map[string]Node   `json:"value,omitempty"`
	lastNodes map[Node]struct{} // TODO : interface like key, is it ok?
	*nestedNode
}

func newAlternation(variants []Node) *alternation {
	n := new(alternation)
	n.Value = make(map[string]Node, len(variants))
	n.lastNodes = make(map[Node]struct{}, len(variants))
	n.nestedNode = newNestedNode()

	variantKey := bytes.NewBuffer(nil)

	for _, variant := range variants {
		variant.Traverse(func(x Node) {
			variantKey.WriteString(x.GetKey())

			if len(x.GetNestedNodes()) == 0 {
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

func (n *alternation) GetKey() string {
	variantKeys := make([]string, 0, len(n.Value))

	for _, variant := range n.Value {
		variantKeys = append(variantKeys, variant.GetKey())
	}

	return strings.Join(variantKeys, ",")
}

func (n *alternation) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Value {
		x.Traverse(f)
	}
}

// TODO : check it without groups too
func (n *alternation) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	n.scanVariants(
		output,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			output.Yield(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(output, input, vTo+1, to, onMatch)
		},
	)
}

func (n *alternation) scanAlternation(
	output Output,
	input TextBuffer,
	from, to int,
	onMatch Callback,
) {
	n.scanVariants(
		output,
		input,
		from,
		to,
		func(variant Node, vFrom, vTo int, empty bool) {
			if _, exists := n.lastNodes[variant]; exists {
				onMatch(variant, vFrom, vTo, empty)
			}
		},
	)
}

func (n *alternation) scanVariants(output Output, input TextBuffer, from, to int, onMatch Callback) {
	position := output.Position()

	for _, variant := range n.Value {
		variant.Scan(output, input, from, to, onMatch)
		output.Rewind(position)
	}
}

type group struct {
	// TODO : it's not really uniq id
	uniqID string
	Value  *alternation `json:"value,omitempty"`
	*nestedNode
}

func (n *group) GetKey() string {
	return fmt.Sprintf("(%s)", n.Value.GetKey())
}

func (n *group) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *group) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	output.Groups().From(n.uniqID, from)

	n.Value.scanAlternation(
		output,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			output.Groups().To(n.uniqID, vTo)
			output.Yield(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(output, input, vTo+1, to, onMatch)
		},
	)

	output.Groups().Delete(n.uniqID)
}

type namedGroup struct {
	Name  string       `json:"name,omitempty"`
	Value *alternation `json:"value,omitempty"`
	*nestedNode
}

func (n *namedGroup) GetKey() string {
	return fmt.Sprintf("(?<%s>%s)", n.Name, n.Value.GetKey())
}

func (n *namedGroup) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *namedGroup) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	output.NamedGroups().From(n.Name, from)
	n.Value.scanAlternation(
		output,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			output.NamedGroups().To(n.Name, vTo)
			output.Yield(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(output, input, vTo+1, to, onMatch)
		},
	)
	output.NamedGroups().Delete(n.Name)
}

type notCapturedGroup struct {
	Value *alternation `json:"value,omitempty"`
	*nestedNode
}

func (n *notCapturedGroup) GetKey() string {
	return fmt.Sprintf("(?:%s)", n.Value.GetKey())
}

func (n *notCapturedGroup) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *notCapturedGroup) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	n.Value.scanAlternation(
		output,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			output.Yield(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(output, input, vTo+1, to, onMatch)
		},
	)
}

type char struct {
	Value rune `json:"value,omitempty"`
	*nestedNode
}

func (n *char) GetKey() string {
	return string(n.Value)
}

func (n *char) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *char) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) == n.Value {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

// add something to empty json value, and in another spec symbols
type dot struct {
	*nestedNode
}

func (n *dot) GetKey() string {
	return "."
}

func (n *dot) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *dot) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) != '\n' {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

type digit struct {
	*nestedNode
}

func (n *digit) GetKey() string {
	return "\\d"
}

func (n *digit) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *digit) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsDigit(x) {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

type nonDigit struct {
	*nestedNode
}

func (n *nonDigit) GetKey() string {
	return "\\D"
}

func (n *nonDigit) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *nonDigit) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsDigit(x) {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

type word struct {
	*nestedNode
}

func (n *word) GetKey() string {
	return "\\w"
}

func (n *word) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *word) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x) {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

type nonWord struct {
	*nestedNode
}

func (n *nonWord) GetKey() string {
	return "\\W"
}

func (n *nonWord) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *nonWord) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !(x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)) {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

type space struct {
	*nestedNode
}

func (n *space) GetKey() string {
	return "\\s"
}

func (n *space) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *space) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsSpace(x) {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

type nonSpace struct {
	*nestedNode
}

func (n *nonSpace) GetKey() string {
	return "\\S"
}

func (n *nonSpace) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *nonSpace) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsSpace(x) {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)
		output.Rewind(pos)
	}
}

type startOfLine struct {
	*nestedNode
}

func (n *startOfLine) GetKey() string {
	return "^"
}

func (n *startOfLine) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *startOfLine) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	// TODO : precache new line positions in buffer?

	if from == 0 || n.isEndOfLine(input, from-1) { // TODO : check \n\r too
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(output, input, from, to, onMatch)
		output.Rewind(pos)
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

func (n *endOfLine) GetKey() string {
	return "$"
}

func (n *endOfLine) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *endOfLine) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	// TODO : precache new line positions in buffer?

	if n.isEndOfLine(input, from) {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(output, input, from, to, onMatch)
		output.Rewind(pos)
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

func (n *startOfString) GetKey() string {
	return "\\A"
}

func (n *startOfString) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *startOfString) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from == 0 {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(output, input, from, to, onMatch)
		output.Rewind(pos)
	}
}

type endOfString struct {
	*nestedNode
}

func (n *endOfString) GetKey() string {
	return "\\z"
}

func (n *endOfString) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *endOfString) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from == input.Size() {
		pos := output.Position()
		output.Yield(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(output, input, from, to, onMatch)
		output.Rewind(pos)
	}
}

// https://www.regular-expressions.info/repeat.html

type quantifier struct {
	From  int  `json:"from"`
	To    *int `json:"to,omitempty"`
	More  bool `json:"more,omitempty"`
	Value Node `json:"value,omitempty"`
	*nestedNode
}

func (n *quantifier) GetKey() string {
	return n.Value.GetKey() + n.getQuantifierKey()
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

func (n *quantifier) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *quantifier) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	start := output.Position()

	n.recursiveScan(1, output, input, from, to, func(_ Node, _, mTo int, empty bool) {
		pos := output.Position()
		output.Yield(n, from, mTo, n.IsEnd(), false)
		onMatch(n, from, mTo, empty)
		n.nestedNode.Match(output, input, mTo+1, to, onMatch)
		output.Rewind(pos)
	})

	output.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.From == 0 {
		if m, exists := output.LastMatchSpan(); exists {
			// TODO : remove condition and this line?
			output.Yield(n, m.To(), m.To(), n.IsEnd(), false)
		} else {
			output.Yield(n, from, from, n.IsEnd(), true)
		}

		n.nestedNode.Match(output, input, from, to, onMatch)
	}

	output.Rewind(start)
}

func (n *quantifier) recursiveScan(
	count int,
	output Output,
	input TextBuffer,
	from, to int,
	onMatch Callback,
) {
	n.Value.Scan(output, input, from, to, func(match Node, mFrom, mTo int, empty bool) {
		if n.To == nil || *n.To >= count {
			if n.inBounds(count) {
				onMatch(match, mFrom, mTo, empty)
			}

			next := count + 1

			if n.To == nil || *n.To >= next {
				n.recursiveScan(next, output, input, mTo+1, to, onMatch)
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

func (n *characterClass) GetKey() string {
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

func (n *characterClass) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *characterClass) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	// TODO : always only one character?
	if unicode.In(x, n.table) {
		pos := output.Position()

		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)

		output.Rewind(pos)
	}
}

type negatedCharacterClass struct {
	table *unicode.RangeTable
	*nestedNode
}

func (n *negatedCharacterClass) GetKey() string {
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

func (n *negatedCharacterClass) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *negatedCharacterClass) Scan(output Output, input TextBuffer, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	// TODO : always only one character?
	if !unicode.In(x, n.table) {
		pos := output.Position()

		output.Yield(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(output, input, from+1, to, onMatch)

		output.Rewind(pos)
	}
}
