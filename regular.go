package regular

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	c "github.com/okneniz/parsec/common"
)

type node interface {
	getKey() string
	getExpressions() dict
	getNestedNodes() index
	addExpression(string)
	isEnd() bool
	// scan()
}

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

		exp, err := defaultParser(buf)
		if err != nil {
			return err
		}

		t.addExpression(str, exp)
	}

	return nil
}

func (t *trie) MarshalJSON() ([]byte, error) {
	output := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	err := encoder.Encode(t.nodes)
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

func (t *trie) String() string {
	data, err := t.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(data)
}

func (t *trie) addExpression(str string, exp expression) {
	ix := t.nodes
	lastNode := exp[len(exp)-1]

	for _, n := range exp {
		key := n.getKey()

		if prev, exists := ix[key]; exists {
			ix = prev.getNestedNodes()
			lastNode = prev
		} else {
			ix[key] = n
			ix = n.getNestedNodes()
			lastNode = n
		}
	}

	lastNode.addExpression(str)
}

// is (foo|bar) is equal (bar|foo) ?
// (fo|f)(o|oo)

type group struct {
	Value       []expression `json:"value,omitempty"`
	Expressions dict         `json:"expression,omitempty"`
	Nested      index        `json:"nested,omitempty"`
}

func (n *group) getKey() string {
	subKeys := make([]string, len(n.Value))

	var b strings.Builder

	for i, exp := range n.Value {
		for _, n := range exp {
			b.WriteString(n.getKey())
		}

		subKeys[i] = b.String()
		b.Reset()
	}

	// TODO : may be sort? order is important?

	x := strings.Join(subKeys, "|")
	return fmt.Sprintf("(%s)", x)
}

func (n *group) getNestedNodes() index {
	return n.Nested
}

func (n *group) getExpressions() dict {
	return n.Expressions
}

func (n *group) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
	}
}

func (n *group) isEnd() bool {
	return len(n.Expressions) == 0
}

type namedGroup struct {
	Name        string       `json:"name,omitempty"`
	Value       []expression `json:"value,omitempty"`
	Expressions dict         `json:"expressions,omitempty"`
	Nested      index        `json:"nested,omitempty"`
}

func (n *namedGroup) getKey() string {
	subKeys := make([]string, len(n.Value))

	var b strings.Builder

	for i, exp := range n.Value {
		for _, n := range exp {
			b.WriteString(n.getKey())
		}

		subKeys[i] = b.String()
		b.Reset()
	}

	// TODO : may be sort? order is important?

	x := strings.Join(subKeys, "|")
	return fmt.Sprintf("(?<%s>%s)", n.Name, x)
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

func (n *namedGroup) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
	}
}

type notCapturedGroup struct {
	Value       []expression `json:"value,omitempty"`
	Expressions dict         `json:"expressions,omitempty"`
	Nested      index        `json:"nested,omitempty"`
}

func (n *notCapturedGroup) getKey() string {
	subKeys := make([]string, len(n.Value))

	var b strings.Builder

	for i, exp := range n.Value {
		for _, n := range exp {
			b.WriteString(n.getKey())
		}

		subKeys[i] = b.String()
		b.Reset()
	}

	// TODO : may be sort? order is important?

	x := strings.Join(subKeys, "|")
	return fmt.Sprintf("(?:%s)", x)
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

func (n *notCapturedGroup) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
	}
}

type char struct {
	Value       string `json:"value,omitempty"`
	Expressions dict   `json:"expressions,omitempty"`
	Nested      index  `json:"nested,omitempty"`
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

func (n *char) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
	}
}

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

func (n *dot) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *digit) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *nonDigit) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *word) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *nonWord) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *space) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *nonSpace) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *startOfLine) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *endOfLine) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
	}
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

func (n *startOfString) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *endOfString) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *rangeNode) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *quantifier) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *positiveSet) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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

func (n *negativeSet) addExpression(str string) {
	if n.Expressions == nil {
		n.Expressions = make(dict)
	}

	if _, exists := n.Expressions[str]; !exists {
		n.Expressions[str] = struct{}{}
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
func newBuffer(str string) c.Buffer[rune, int] {
	b := new(simpleBuffer)
	b.data = []rune(str)
	b.position = 0
	return b
}

type Trie interface {
	Add(...string) error
	// IsInclude(string) bool
	// Match(string) (MatchedData, error)
	// IsMatched(string) (bool, error)
}

type MatchedData interface {
	From() int
	To() int
	String() string
	Groups() map[string]string
}

type expression = []node
type index map[string]node
type dict map[string]struct{}

type parser = c.Combinator[rune, int, node]
type expressionParser = c.Combinator[rune, int, expression]

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


func parseRegexp(except ...rune) expressionParser {
	var nestedRegexp expressionParser

	sep := c.Eq[rune, int]('|')

	// TODO : union without groups?
	union := func(buf c.Buffer[rune, int]) ([]expression, error) {
		result := make([]expression, 0, 1)

		variant, err := nestedRegexp(buf)
		if err != nil {
			return result, nil
		}

		result = append(result, variant)

		for !buf.IsEOF() {
			pos := buf.Position()

			_, err = sep(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			variant, err = nestedRegexp(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			result = append(result, variant)
		}

		return result, nil
	}

	groups := choice(
		parseNotCapturedGroup(union),
		parseNamedGroup(union),
		parseGroup(union),
	)

	characters := choice(
		parseInvalidQuantifier(),
		parseMetaCharacters(),
		parseSymbols(),
		parseEscapedMetacharacters(),
		parseCharacter(except...),
	)

	// TODO : return error for invalid escaped chars like '\x' (check on rubular)

	// is it possible for nested set?
	setsCombinatrors := choice(
		parseRange(append(except, ']')...),
		parseMetaCharacters(),
		parseEscapedMetacharacters(),
		parseCharacter(append(except, ']')...),
	)

	sets := choice(
		parseNegativeSet(setsCombinatrors),
		parsePositiveSet(setsCombinatrors),
	)

	regexp := c.Some(
		1,
		parseOptionalQuantifier(
			choice(
				sets,
				groups,
				characters,
			),
		),
	)

	if len(except) != 0 {
		nestedRegexp = regexp
	} else {
		nestedRegexp = parseRegexp(append(except, ')', '|')...)
	}

	nestedRegexp = c.Try(nestedRegexp)

	return regexp
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

func braces[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return between(
		c.Eq[rune, int]('{'),
		body,
		c.Eq[rune, int]('}'),
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

func parseEscapedMetacharacters() parser {
	chars := ".?+*^$[]{}()"
	parsers := make([]parser, len(chars))

	for i, c := range chars {
		parsers[i] = parseEscapedMetacharacter(c) // todo - speed up it - use one parser without try
	}

	return choice(parsers...)
}

func parseEscapedMetacharacter(value rune) parser {
	str := string([]rune{'\\', value})
	parse := SkipString(str)

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := char{
			Value:  str,
			Nested: make(index, 0),
		}

		return &x, nil
	}
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
	digit := c.Try(number())
	lookup := c.Satisfy[rune, int](false, c.Anything[rune])
	skip := c.Any[rune, int]()

	parseQuantifier := c.Try(
		c.MapAs(
			map[rune]c.Combinator[rune, int, quantifier]{
				'?': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf) // TODO : remove skip? and lookup?
					if err != nil {
						return q, err
					}

					to := 1
					q.From = 0
					q.To = &to
					q.More = false

					return q, nil
				},
				'+': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf)
					if err != nil {
						return q, err
					}

					q.From = 1
					q.More = true

					return q, nil
				},
				'*': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf)
					if err != nil {
						return q, err
					}

					q.From = 0
					q.More = true

					return q, nil
				},
				'{': braces[quantifier](func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					from, err := digit(buf)
					if err != nil {
						return q, err
					}

					q.From = from

					x, err := lookup(buf)
					if err != nil {
						return q, nil
					}
					if x != ',' {
						return q, nil
					}
					_, err = skip(buf)
					if err != nil {
						return q, err
					}

					q.More = true

					to, err := digit(buf)
					if err != nil {
						return q, err
					}

					q.To = &to

					return q, nil
				},
				),
			},
			lookup,
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
			Value:  string(c),
			Nested: make(index, 0),
		}

		return &x, nil
	}
}

func parseSymbols() parser {
	return c.MapAs(
		map[rune]c.Combinator[rune, int, node]{
			'.': func(buf c.Buffer[rune, int]) (node, error) {
				x := dot{
					Nested: make(index, 0),
				}

				return &x, nil
			},
			'^': func(buf c.Buffer[rune, int]) (node, error) {
				x := startOfLine{
					Nested: make(index, 0),
				}

				return &x, nil
			},
			'$': func(buf c.Buffer[rune, int]) (node, error) {
				x := endOfLine{
					Nested: make(index, 0),
				}

				return &x, nil
			},
		},
		c.Any[rune, int](),
	)
}

func parseMetaCharacters() parser {
	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			map[rune]c.Combinator[rune, int, node]{
				'd': func(buf c.Buffer[rune, int]) (node, error) {
					x := digit{
						Nested: make(index, 0),
					}

					return &x, nil
				},
				'D': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonDigit{
						Nested: make(index, 0),
					}

					return &x, nil
				},
				'w': func(buf c.Buffer[rune, int]) (node, error) {
					x := word{
						Nested: make(index, 0),
					}

					return &x, nil
				},
				'W': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonWord{
						Nested: make(index, 0),
					}

					return &x, nil
				},
				's': func(buf c.Buffer[rune, int]) (node, error) {
					x := space{
						Nested: make(index, 0),
					}

					return &x, nil
				},
				'S': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonSpace{
						Nested: make(index, 0),
					}

					return &x, nil
				},
				'A': func(buf c.Buffer[rune, int]) (node, error) {
					x := startOfString{
						Nested: make(index, 0),
					}

					return &x, nil
				},
				'z': func(buf c.Buffer[rune, int]) (node, error) {
					x := endOfString{
						Nested: make(index, 0),
					}

					return &x, nil
				},
			},
			c.Any[rune, int](),
		),
	)
}

func parseGroup(union c.Combinator[rune, int, []expression]) parser {
	return parens(
		func(buf c.Buffer[rune, int]) (node, error) {
			variants, err := union(buf)
			if err != nil {
				return nil, err
			}

			x := group{
				Value:  variants,
				Nested: make(index, 0),
			}

			return &x, nil
		},
	)
}

func parseNotCapturedGroup(union c.Combinator[rune, int, []expression]) parser {
	before := SkipString("?:")

	return parens(
		func(buf c.Buffer[rune, int]) (node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			variants, err := union(buf)
			if err != nil {
				return nil, err
			}

			x := notCapturedGroup{
				Value:  variants,
				Nested: make(index, 0),
			}

			return &x, nil
		},
	)
}

func parseNamedGroup(union c.Combinator[rune, int, []expression], except ...rune) parser {
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

			variants, err := union(buf)
			if err != nil {
				return nil, err
			}

			x := namedGroup{
				Name:   string(name),
				Value:  variants,
				Nested: make(index, 0),
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
			Value:  set,
			Nested: make(index, 0),
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
			Value:  set,
			Nested: make(index, 0),
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
