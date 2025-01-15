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

	// Add parent to travers
	// Should return bool to interupt traversing?
	Traverse(func(Node))

	// TODO : works only for fixed chain with one end node?
	// don't work for tree?

	// TODO : it's improtanto for group too have chain in Value instead tree
	// make special type for this case?

	// TODO : what about alternation of chains?
	Size() (int, bool)

	// TODO : what about anchors, is it endless or zero sized?
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
	for k, nested := range n.Nested {
		// pos := scanner.Position()
		// groupsPos := scanner.GroupsPosition()
		// namedGroupPos := scanner.NamedGroupsPosition()

		fmt.Println("scan nested", k, from, to)
		nested.Visit(scanner, input, from, to, onMatch)

		// scanner.Rewind(pos)
		// scanner.RewindGroups(groupsPos)
		// scanner.RewindNamedGroups(namedGroupPos)
	}
}

func (n *nestedNode) NestedSize() (int, bool) {
	if len(n.Nested) == 0 {
		return 0, true
	}

	var size *int

	for _, child := range n.Nested {
		if x, fixedSize := child.Size(); fixedSize {
			if size != nil && *size != x {
				return 0, false
			}

			size = &x
		} else {
			return 0, false
		}
	}

	if size == nil {
		return 0, false
	}

	return *size, true
}

type alternation struct {
	// TODO : why not list, order is important
	Value     map[string]Node   `json:"value,omitempty"`
	lastNodes map[Node]struct{} // TODO : interface like key, is it ok?
	size      *int
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

func (n *alternation) Size() (int, bool) {
	var size *int
	for _, variant := range n.Value {
		if x, fixedSize := variant.Size(); fixedSize {
			if size != nil && *size != x {
				return 0, false
			}

			size = &x
		} else {
			return 0, false
		}
	}

	if size == nil {
		return 0, false
	}

	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return *size + nestedSize, true
	}

	return 0, false
}

type group struct {
	Value *alternation `json:"value,omitempty"`
	*nestedNode
}

func newGroup(expression *alternation) *group {
	g := &group{
		nestedNode: newNestedNode(),
		Value:      expression,
	}

	return g
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

func (n *group) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

type namedGroup struct {
	Name  string       `json:"name,omitempty"`
	Value *alternation `json:"value,omitempty"`
	*nestedNode
}

func newNamedGroup(name string, expression *alternation) *namedGroup {
	g := &namedGroup{
		Name:       name,
		nestedNode: newNestedNode(),
		Value:      expression,
	}

	return g
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

func (n *namedGroup) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

// TODO : what about back references?
// Add tests too

type notCapturedGroup struct {
	Value *alternation `json:"value,omitempty"`
	*nestedNode
}

func newNotCapturedGroup(expression *alternation) *notCapturedGroup {
	g := &notCapturedGroup{
		Value:      expression,
		nestedNode: newNestedNode(),
	}

	return g
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

func (n *notCapturedGroup) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

// not simple node with table because §diferent behaviour for different scan options
// TODO : add something to empty json value, and in another spec symbols
type dot struct {
	size *int
	*nestedNode
}

func newDot() *dot {
	return &dot{
		nestedNode: newNestedNode(),
	}
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

func (n *dot) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
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

func (n *startOfLine) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
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

func (n *endOfLine) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
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

func (n *startOfString) Size() (int, bool) {
	return 0, true
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

func (n *endOfString) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
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

// TODO : add tests to fail on parsing not fixed size quantificators in look behind assertions
func (n *quantifier) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
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

func (n *simpleNode) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	fmt.Println("simpleNode without fixed size", n)

	return 0, false
}

// back reference \1, \2 or \9

type referenceNode struct {
	key   string
	index int
	*nestedNode
}

func nodeForReference(index int) *referenceNode {
	return &referenceNode{
		key:        fmt.Sprintf("\\%d", index),
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

func (n *referenceNode) Size() (int, bool) {
	return 0, false
}

// named back reference \k<name>

type nameReferenceNode struct {
	key  string
	name string
	*nestedNode
}

func nodeForNameReference(name string) *nameReferenceNode {
	return &nameReferenceNode{
		key:        fmt.Sprintf("\\k<%s>", name),
		name:       name,
		nestedNode: newNestedNode(),
	}
}

func (n *nameReferenceNode) GetKey() string {
	return n.key
}

func (n *nameReferenceNode) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *nameReferenceNode) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	matchSpan, exists := scanner.GetNamedGroup(n.name)

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

func (n *nameReferenceNode) Size() (int, bool) {
	return 0, false
}

type LookAhead struct {
	Value *alternation `json:"value,omitempty"`
	*nestedNode
}

func newLookAhead(expression *alternation) *LookAhead {
	return &LookAhead{
		Value:      expression,
		nestedNode: newNestedNode(),
	}
}

func (n *LookAhead) GetKey() string {
	return fmt.Sprintf("(?=%s)", n.Value.GetKey())
}

func (n *LookAhead) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *LookAhead) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.Value.visitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			holesPos := scanner.HolesPosition()

			// what about empty spans, just skip it?
			scanner.MarkAsHole(from, vTo) // or just scanner rewind to "FROM" pos without holes?
			scanner.Match(n, from, from, n.IsEnd(), true)
			onMatch(n, from, from, true)

			scanner.RewindHoles(holesPos)
			n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

			scanner.Rewind(pos)
		},
	)
}

func (n *LookAhead) Size() (int, bool) {
	return 0, false
}

type NegativeLookAhead struct {
	Value *alternation `json:"value,omitempty"`
	*nestedNode
}

func newNegativeLookAhead(expression *alternation) *NegativeLookAhead {
	return &NegativeLookAhead{
		Value:      expression,
		nestedNode: newNestedNode(),
	}
}

func (n *NegativeLookAhead) GetKey() string {
	return fmt.Sprintf("(?!%s)", n.Value.GetKey())
}

func (n *NegativeLookAhead) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *NegativeLookAhead) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	matched := false
	pos := scanner.Position()

	n.Value.visitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			matched = true
			scanner.Rewind(pos)
			// TODO : stop here
		},
	)

	scanner.Rewind(pos)

	if !matched {
		scanner.Rewind(pos)

		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)

		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(pos)
}

func (n *NegativeLookAhead) Size() (int, bool) {
	return 0, false
}

type LookBehind struct {
	Value             *alternation `json:"value,omitempty"`
	subExpressionSize int
	*nestedNode
}

func newLookBehind(expression *alternation) (*LookBehind, error) {
	size, fixedSize := expression.Size()
	if !fixedSize {
		return nil, fmt.Errorf("Invalid pattern in look-behind, must be fixed size")
	}

	return &LookBehind{
		Value:             expression,
		subExpressionSize: size,
		nestedNode:        newNestedNode(),
	}, nil
}

func (n *LookBehind) GetKey() string {
	return fmt.Sprintf("(?<=%s)", n.Value.GetKey())
}

func (n *LookBehind) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *LookBehind) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	// TODO : what about anchors?
	if from < n.subExpressionSize {
		return
	}

	pos := scanner.Position()

	n.Value.visitAlternation(
		scanner,
		input,
		from-n.subExpressionSize,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.Rewind(pos)

			scanner.Match(n, from, from, n.IsEnd(), true)
			onMatch(n, from, from, true)
			n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

			scanner.Rewind(pos)
		},
	)
}

func (n *LookBehind) Size() (int, bool) {
	return 0, false
}

type NegativeLookBehind struct {
	Value             *alternation `json:"value,omitempty"`
	subExpressionSize int
	*nestedNode
}

func newNegativeLookBehind(expression *alternation) (*NegativeLookBehind, error) {
	size, fixedSize := expression.Size()
	if !fixedSize {
		return nil, fmt.Errorf("Invalid pattern in negative look-behind, must be fixed size")
	}

	return &NegativeLookBehind{
		Value:             expression,
		subExpressionSize: size,
		nestedNode:        newNestedNode(),
	}, nil
}

func (n *NegativeLookBehind) GetKey() string {
	return fmt.Sprintf("(?<!%s)", n.Value.GetKey())
}

func (n *NegativeLookBehind) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *NegativeLookBehind) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	// TODO : what about anchors?
	pos := scanner.Position()

	if from < n.subExpressionSize {
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
		return
	}

	matched := false
	fmt.Println("debug nlb :", from, to)

	n.Value.visitAlternation(
		scanner,
		input,
		from-n.subExpressionSize,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.Rewind(pos)
			matched = true
			// TODO : stop here
		},
	)

	fmt.Println("debug nlb =", from, to, matched)

	scanner.Rewind(pos)

	if !matched {
		scanner.Rewind(pos)
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(pos)
}

func (n *NegativeLookBehind) Size() (int, bool) {
	return 0, false
}
