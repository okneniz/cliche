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

	Visit(Scanner, Input, int, int, Callback)
	Merge(Node)
	Traverse(func(Node))
}

type Callback func(x Node, from int, to int, empty bool)

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
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	pos := scanner.Position()

	for _, nested := range n.Nested {
		nested.Visit(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
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

// In the tutorial topic about alternation, I explained that the regex engine will stop
// as soon as it finds a matching alternative.
// The POSIX standard, however, mandates that the longest match be returned.
// When applying Set|SetValue to SetValue, a POSIX-compliant regex engine will
// match SetValue entirely.
// Even if the engine is a regex-directed NFA engine, POSIX requires that it
// simulates DFA text-directed matching by trying all alternatives,
// and returning the longest match, in this case SetValue.
// A traditional NFA engine would match Set, as do all other regex flavors discussed
// on this website.

// A POSIX-compliant engine will still find the leftmost match.
// If you apply Set|SetValue to Set or SetValue once, it will match Set.
// The first position in the string is the leftmost position where our regex can find a
//  valid match.
// The fact that a longer match can be found further in the string is irrelevant.
// If you apply the regex a second time, continuing at the first space in the string,
// then SetValue will be matched.
// A traditional NFA engine would match Set at the start of the string as the first match,
// and Set at the start of the 3rd word in the string as the second match.

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
func (n *alternation) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.scanVariants(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(scanner, input, vTo+1, to, onMatch)
		},
	)
}

func (n *alternation) scanAlternation(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	n.scanVariants(
		scanner,
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

func (n *alternation) scanVariants(scanner Scanner, input Input, from, to int, onMatch Callback) {
	position := scanner.Position()

	for _, variant := range n.Value {
		variant.Visit(scanner, input, from, to, onMatch)
		scanner.Rewind(position)
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

func (n *group) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	scanner.StartGroup(n.uniqID, from)

	n.Value.scanAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.EndGroup(n.uniqID, vTo)
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(scanner, input, vTo+1, to, onMatch)
		},
	)

	scanner.DeleteGroup(n.uniqID)
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

func (n *namedGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	scanner.StartNamedGroup(n.Name, from)

	n.Value.scanAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.EndNamedGroup(n.Name, vTo)
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(scanner, input, vTo+1, to, onMatch)
		},
	)

	scanner.DeleteNamedGroup(n.Name)
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

func (n *notCapturedGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.Value.scanAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.Match(scanner, input, vTo+1, to, onMatch)
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

func (n *char) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) == n.Value {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *dot) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) != '\n' {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *digit) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsDigit(x) {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *nonDigit) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsDigit(x) {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *word) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x) {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *nonWord) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !(x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)) {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *space) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.IsSpace(x) {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *nonSpace) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.IsSpace(x) {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)
		scanner.Rewind(pos)
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

func (n *startOfLine) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	// TODO : precache new line positions in buffer?

	if from == 0 || n.isEndOfLine(input, from-1) { // TODO : check \n\r too
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
	}
}

func (n *startOfLine) isEndOfLine(input Input, idx int) bool {
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

func (n *endOfLine) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	// TODO : precache new line positions in buffer?

	if n.isEndOfLine(input, from) {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
	}
}

// TODO : check \n\r too
func (n *endOfLine) isEndOfLine(input Input, idx int) bool {
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

func (n *startOfString) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from == 0 {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
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

func (n *endOfString) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from == input.Size() {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.Match(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
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
	} else if n.To != nil && n.From != *n.To {
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

func (n *quantifier) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	start := scanner.Position()

	n.recursiveVisit(1, scanner, input, from, to, func(_ Node, _, mTo int, empty bool) {
		pos := scanner.Position()
		scanner.Match(n, from, mTo, n.IsEnd(), false)
		onMatch(n, from, mTo, empty)
		n.nestedNode.Match(scanner, input, mTo+1, to, onMatch)
		scanner.Rewind(pos)
	})

	scanner.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.From == 0 {
		scanner.Match(n, from, from, n.IsEnd(), true)
		n.nestedNode.Match(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(start)
}

func (n *quantifier) recursiveVisit(
	count int,
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	n.Value.Visit(scanner, input, from, to, func(match Node, mFrom, mTo int, empty bool) {
		if n.To == nil || *n.To >= count {
			if n.inBounds(count) {
				onMatch(match, mFrom, mTo, empty)
			}

			next := count + 1

			if n.To == nil || *n.To >= next {
				n.recursiveVisit(next, scanner, input, mTo+1, to, onMatch)
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

func (n *characterClass) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if unicode.In(x, n.table) {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
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

func (n *negatedCharacterClass) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if !unicode.In(x, n.table) {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
	}
}

type bracket struct {
	key       string
	predicate func(rune) bool
	*nestedNode
}

func (n *bracket) GetKey() string {
	return n.key
}

func (n *bracket) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *bracket) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)

	if n.predicate(x) {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.Match(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
	}
}
