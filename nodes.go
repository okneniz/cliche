package cliche

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

type Node interface {
	GetKey() string
	GetExpressions() Set
	AddExpression(string)
	GetNestedNodes() map[string]Node
	IsEnd() bool

	Visit(Scanner, Input, int, int, Callback)
	Merge(Node) // remove, implement Merge(Node, Node) method in parser or tree
	Traverse(func(Node))
}

type Callback func(x Node, from int, to int, empty bool)

type nestedNode struct {
	Expressions Set             `json:"expressions,omitempty"`
	Nested      map[string]Node `json:"nested,omitempty"`
}

func newNestedNode() *nestedNode {
	n := new(nestedNode)
	n.Nested = make(map[string]Node)
	return n
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

func (n *nestedNode) VisitNested(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	for _, nested := range n.Nested {
		// pos := scanner.Position()
		// groupsPos := scanner.GroupsPosition()
		// namedGroupPos := scanner.NamedGroupsPosition()

		nested.Visit(scanner, input, from, to, onMatch)

		// scanner.Rewind(pos)
		// scanner.RewindGroups(groupsPos)
		// scanner.RewindNamedGroups(namedGroupPos)
	}
}

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
	n.visitVariants(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.VisitNested(scanner, input, vTo+1, to, onMatch)
			scanner.Rewind(pos)
		},
	)
}

func (n *alternation) visitAlternation(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	n.visitVariants(
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

func (n *alternation) visitVariants(scanner Scanner, input Input, from, to int, onMatch Callback) {
	for _, variant := range n.Value {
		variant.Visit(scanner, input, from, to, onMatch)
	}
}

type group struct {
	Value *alternation `json:"value,omitempty"`
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
	n.Value.visitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			groupsPos := scanner.GroupsPosition()

			// TODO : why to? what about empty captures
			scanner.MatchGroup(from, vTo)
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindGroups(groupsPos)
		},
	)
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
	n.Value.visitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			groupsPos := scanner.NamedGroupsPosition()

			// TODO : why to? what about empty captures
			scanner.MatchNamedGroup(n.Name, from, vTo)
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindNamedGroups(groupsPos)
		},
	)
}

// TODO : what about back references?
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
	n.Value.visitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()

			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
		},
	)
}

// not simple node with table because diferent behaviour for different scan options
// TODO : add something to empty json value, and in another spec symbols
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
		n.nestedNode.VisitNested(scanner, input, from+1, to, onMatch)

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

	if from == 0 || n.isEndOfLine(input, from-1) { // TODO : check \n\r too
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

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
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

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
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
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
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
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
		n.nestedNode.VisitNested(scanner, input, mTo+1, to, onMatch)
		scanner.Rewind(pos)
	})

	scanner.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.From == 0 {
		scanner.Match(n, from, from, n.IsEnd(), true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
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

type simpleNode struct {
	key       string
	predicate func(rune) bool
	*nestedNode
}

func nodeForTable(table *unicode.RangeTable) *simpleNode {
	return &simpleNode{
		key: rangeTableKey(table),
		predicate: func(r rune) bool {
			return unicode.In(r, table)
		},
		nestedNode: newNestedNode(),
	}
}

func nodeForChar(r rune) *simpleNode {
	table := rangetable.New(r)
	return nodeForTable(table)
}

func (n *simpleNode) GetKey() string {
	return n.key
}

func (n *simpleNode) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *simpleNode) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if n.predicate(input.ReadAt(from)) {
		pos := scanner.Position()
		groupsPos := scanner.GroupsPosition()

		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.VisitNested(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
		scanner.RewindGroups(groupsPos)
	}
}

// back references \1, \2 or \9

type referenceNode struct {
	key   string
	index int
	*nestedNode
}

func nodeForReference(index int) *referenceNode {
	return &referenceNode{
		key:        fmt.Sprintf("references to \\%d", index),
		index:      index,
		nestedNode: newNestedNode(),
	}
}

func (n *referenceNode) GetKey() string {
	return n.key
}

func (n *referenceNode) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *referenceNode) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	matchSpan, exists := scanner.GetGroup(n.index)

	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Regular_expressions/Backreference
	//
	// If the referenced capturing group is unmatched (for example, because it belongs to an unmatched alternative in a disjunction),
	// or the group hasn't matched yet (for example, because it lies to the right of the backreference),
	// the backreference always succeeds (as if it matches the empty string).

	pos := scanner.Position()

	if !exists || matchSpan.Empty() {
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

		scanner.Rewind(pos)
	} else {
		// TODO : what about empty matches?

		current := from

		// match the same string
		for prev := matchSpan.From(); prev <= matchSpan.To(); prev++ {
			if current >= input.Size() {
				scanner.Rewind(pos)
				return
			}

			expected := input.ReadAt(prev)
			actual := input.ReadAt(current)

			if expected != actual {
				scanner.Rewind(pos)
				return
			}

			current++
		}

		scanner.Match(n, from, current-1, n.IsEnd(), false)
		onMatch(n, from, current-1, false)

		n.nestedNode.VisitNested(scanner, input, current, to, onMatch)
		scanner.Rewind(pos)
	}
}
