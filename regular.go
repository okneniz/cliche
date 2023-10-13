package regular

import (
	"errors"

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

func (t *trie) addExpression(str string, exp expression) {
	ix := t.nodes
	l := len(exp)

	for i, n := range exp {
		key := n.getKey()

		if prev, exists := ix[key]; exists {
			ix = prev.getNestedNodes()
		} else {
			ix[key] = n
			ix = n.getNestedNodes()
		}

		if i == l {
			ix[key].addExpression(str)
		}
	}
}

// is (foo|bar) is equal (bar|foo) ?
// (fo|f)(o|oo)

type group struct {
	key         string
	value       []expression
	expressions dict
	nested      index
}

func (n *group) getKey() string {
	return n.key
}

func (n *group) getNestedNodes() index {
	return n.nested
}

func (n *group) getExpressions() dict {
	return n.expressions
}

func (n *group) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

func (n *group) isEnd() bool {
	return len(n.expressions) == 0
}

type namedGroup struct {
	key         string
	name        string
	value       []expression
	expressions dict
	nested      index
}

func (n *namedGroup) getKey() string {
	return n.key
}

func (n *namedGroup) getNestedNodes() index {
	return n.nested
}

func (n *namedGroup) getExpressions() dict {
	return n.expressions
}

func (n *namedGroup) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *namedGroup) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type notCapturedGroup struct {
	key         string
	value       []expression
	expressions dict
	nested      index
}

func (n *notCapturedGroup) getKey() string {
	return n.key
}

func (n *notCapturedGroup) getNestedNodes() index {
	return n.nested
}

func (n *notCapturedGroup) getExpressions() dict {
	return n.expressions
}

func (n *notCapturedGroup) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *notCapturedGroup) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type char struct {
	key         string
	value       rune
	expressions dict
	nested      index
}

func (n *char) getKey() string {
	return n.key
}

func (n *char) getNestedNodes() index {
	return n.nested
}

func (n *char) getExpressions() dict {
	return n.expressions
}

func (n *char) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *char) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type dot struct {
	key         string
	expressions dict
	nested      index
}

func (n *dot) getKey() string {
	return n.key
}

func (n *dot) getNestedNodes() index {
	return n.nested
}

func (n *dot) getExpressions() dict {
	return n.expressions
}

func (n *dot) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *dot) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type digit struct {
	key         string
	expressions dict
	nested      index
}

func (n *digit) getKey() string {
	return n.key
}

func (n *digit) getNestedNodes() index {
	return n.nested
}

func (n *digit) getExpressions() dict {
	return n.expressions
}

func (n *digit) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *digit) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type nonDigit struct {
	key         string
	expressions dict
	nested      index
}

func (n *nonDigit) getKey() string {
	return n.key
}

func (n *nonDigit) getNestedNodes() index {
	return n.nested
}

func (n *nonDigit) getExpressions() dict {
	return n.expressions
}

func (n *nonDigit) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *nonDigit) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type word struct {
	key         string
	expressions dict
	nested      index
}

func (n *word) getKey() string {
	return n.key
}

func (n *word) getNestedNodes() index {
	return n.nested
}

func (n *word) getExpressions() dict {
	return n.expressions
}

func (n *word) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *word) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type nonWord struct {
	key         string
	expressions dict
	nested      index
}

func (n *nonWord) getKey() string {
	return n.key
}

func (n *nonWord) getNestedNodes() index {
	return n.nested
}

func (n *nonWord) getExpressions() dict {
	return n.expressions
}

func (n *nonWord) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *nonWord) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type space struct {
	key         string
	expressions dict
	nested      index
}

func (n *space) getKey() string {
	return n.key
}

func (n *space) getNestedNodes() index {
	return n.nested
}

func (n *space) getExpressions() dict {
	return n.expressions
}

func (n *space) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *space) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type nonSpace struct {
	key         string
	expressions dict
	nested      index
}

func (n *nonSpace) getKey() string {
	return n.key
}

func (n *nonSpace) getNestedNodes() index {
	return n.nested
}

func (n *nonSpace) getExpressions() dict {
	return n.expressions
}

func (n *nonSpace) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *nonSpace) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type startOfLine struct {
	key         string
	expressions dict
	nested      index
}

func (n *startOfLine) getKey() string {
	return n.key
}

func (n *startOfLine) getNestedNodes() index {
	return n.nested
}

func (n *startOfLine) getExpressions() dict {
	return n.expressions
}

func (n *startOfLine) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *startOfLine) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type endOfLine struct {
	key         string
	expressions dict
	nested      index
}

func (n *endOfLine) getKey() string {
	return n.key
}

func (n *endOfLine) getNestedNodes() index {
	return n.nested
}

func (n *endOfLine) getExpressions() dict {
	return n.expressions
}

func (n *endOfLine) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *endOfLine) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type startOfString struct {
	key         string
	expressions dict
	nested      index
}

func (n *startOfString) getKey() string {
	return n.key
}

func (n *startOfString) getNestedNodes() index {
	return n.nested
}

func (n *startOfString) getExpressions() dict {
	return n.expressions
}

func (n *startOfString) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *startOfString) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type endOfString struct {
	key         string
	expressions dict
	nested      index
}

func (n *endOfString) getKey() string {
	return n.key
}

func (n *endOfString) getNestedNodes() index {
	return n.nested
}

func (n *endOfString) getExpressions() dict {
	return n.expressions
}

func (n *endOfString) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *endOfString) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type rangeNode struct {
	key         string
	from        rune
	to          rune
	nested      index
	expressions dict
}

func (n *rangeNode) getKey() string {
	return n.key
}

func (n *rangeNode) getNestedNodes() index {
	return n.nested
}

func (n *rangeNode) getExpressions() dict {
	return n.expressions
}

func (n *rangeNode) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *rangeNode) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type quantifier struct {
	key         string
	from        int
	to          *int
	more        bool
	value       node
	expressions dict
	nested      index
}

func (n *quantifier) getKey() string {
	return n.key
}

func (n *quantifier) getNestedNodes() index {
	return n.nested
}

func (n *quantifier) getExpressions() dict {
	return n.expressions
}

func (n *quantifier) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *quantifier) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type positiveSet struct {
	key         string
	value       []node
	expressions dict
	nested      index
}

func (n *positiveSet) getKey() string {
	return n.key
}

func (n *positiveSet) getNestedNodes() index {
	return n.nested
}

func (n *positiveSet) getExpressions() dict {
	return n.expressions
}

func (n *positiveSet) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *positiveSet) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
	}
}

type negativeSet struct {
	key         string
	value       []node
	expressions dict
	nested      index
}

func (n *negativeSet) getKey() string {
	return n.key
}

func (n *negativeSet) getNestedNodes() index {
	return n.nested
}

func (n *negativeSet) getExpressions() dict {
	return n.expressions
}

func (n *negativeSet) isEnd() bool {
	return len(n.expressions) == 0
}

func (n *negativeSet) addExpression(str string) {
	if n.expressions == nil {
		n.expressions = make(dict)
	}

	if _, exists := n.expressions[str]; !exists {
		n.expressions[str] = struct{}{}
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

// IsEOF - true if buffer ended.
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

type parser = c.Combinator[rune, int, node]
type expression = []node
type expressionParser = c.Combinator[rune, int, expression]

type index map[string]node
type dict map[string]struct{}

var (
	defaultParser = parseRegexp()
	none          = struct{}{}

	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

func New(exps ...string) (Trie, error) {
	t := new(trie)
	t.nodes = make(index, len(exps))

	err := t.Add(exps...)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func parseRegexp(except ...rune) expressionParser {
	var (
		regexp       expressionParser
		nestedRegexp expressionParser
		groups       parser
	)

	union := func(buf c.Buffer[rune, int]) ([]expression, error) {
		result := make([]expression, 0, 1)

		variant, err := nestedRegexp(buf)
		if err != nil {
			return result, nil
		}

		result = append(result, variant)

		for !buf.IsEOF() {
			variant, err = nestedRegexp(buf)
			if err != nil {
				break
			}

			result = append(result, variant)
		}

		return result, nil
	}

	groups = choice(
		parseNotCapturedGroup(union),
		parseNamedGroup(union),
		parseGroup(union),
	)

	characters := choice(
		parseInvalidQuantifier(),
		parseEscapedMetacharacters(),
		parseDot(),
		parseMetaCharacters(),
		parseCharacter(except...),
	)

	setsCombinatrors := choice( // where dot?
		parseRange(append(except, ']')...),
		parseEscapedMetacharacters(),
		parseMetaCharacters(),
		parseCharacter(except...),
	)

	// TODO - improve parsing - use one parser for meta characters - avoid tries

	sets := choice(
		parsePositiveSet(setsCombinatrors),
		parseNegativeSet(setsCombinatrors),
	)

	regexp = c.Some(
		0,
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
			value:  value,
			nested: make(index, 0),
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
		x, err := buf.Read(true)
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
					q.from = 0
					q.to = &to
					q.more = false

					return q, nil
				},
				'+': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf)
					if err != nil {
						return q, err
					}

					q.from = 1
					q.more = true

					return q, nil
				},
				'*': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf)
					if err != nil {
						return q, err
					}

					q.from = 0
					q.more = true

					return q, nil
				},
				'{': braces[quantifier](func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					from, err := digit(buf)
					if err != nil {
						return q, err
					}

					q.from = from

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

					q.more = true

					to, err := digit(buf)
					if err != nil {
						return q, err
					}

					q.to = &to

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

		q.value = x
		q.nested = make(index, 0)

		return &q, nil
	}
}

func SkipString(data string) c.Combinator[rune, int, struct{}] {
	return func(simpleBuffer c.Buffer[rune, int]) (struct{}, error) {
		l := len(data)
		for _, x := range data {
			r, err := simpleBuffer.Read(true)
			if err != nil {
				return none, err
			}
			if x != r {
				return none, c.NothingMatched
			}
			l = -1
		}

		if l != 0 {
			return none, c.NothingMatched
		}

		return none, nil
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
			value:  c,
			nested: make(index, 0),
		}

		return &x, nil
	}
}

func parseDot() parser {
	parse := c.Eq[rune, int]('.')

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := dot{
			nested: make(index, 0),
		}

		return &x, nil
	}
}

func parseMetaCharacters() parser {
	return c.Skip(
		SkipString("\\"),
		c.MapAs(
			map[rune]c.Combinator[rune, int, node]{
				'd': func(buf c.Buffer[rune, int]) (node, error) {
					x := digit{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'D': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonDigit{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'w': func(buf c.Buffer[rune, int]) (node, error) {
					x := word{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'W': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonWord{
						nested: make(index, 0),
					}

					return &x, nil
				},
				's': func(buf c.Buffer[rune, int]) (node, error) {
					x := space{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'S': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonSpace{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'^': func(buf c.Buffer[rune, int]) (node, error) {
					x := startOfLine{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'$': func(buf c.Buffer[rune, int]) (node, error) {
					x := endOfLine{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'A': func(buf c.Buffer[rune, int]) (node, error) {
					x := startOfString{
						nested: make(index, 0),
					}

					return &x, nil
				},
				'z': func(buf c.Buffer[rune, int]) (node, error) {
					x := endOfString{
						nested: make(index, 0),
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
				value:  variants,
				nested: make(index, 0),
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
				value:  variants,
				nested: make(index, 0),
			}

			return &x, nil
		},
	)
}

func parseNamedGroup(union c.Combinator[rune, int, []expression], except ...rune) parser {
	groupName := angles(
		c.Skip(
			c.Eq[rune, int]('?'),
			c.Many(0, c.NoneOf[rune, int](append(except, '>')...)),
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
				name:   string(name),
				value:  variants,
				nested: make(index, 0),
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
			value:  set,
			nested: make(index, 0),
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
			value:  set,
			nested: make(index, 0),
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
			from:   f,
			to:     t,
			nested: make(index, 0),
		}

		return &x, nil
	}
}
