package regular

import (
	"errors"

	c "github.com/okneniz/parsec/common"
)

type node interface {
	getKey() string
	getExpressions() []string // more than rune for quantifiers
	getNestedNodes() index
	isEnd() bool
	// scan()
}

type trie struct {
	nodes index
}

func (t *trie) Add(strs ...string) error {
	for _, str := range strs {
		buf := newBuffer([]rune(str))

		exp, err := defaultParser(buf)
		if err != nil {
			return err
		}

		t.addExpression(exp)
	}

	return nil
}

func (t *trie) addExpression(exp expression) {
	ix := t.nodes

	for _, n := range exp {
		key := n.getKey()

		if prev, exists := ix[key]; exists {
			ix = prev.getNestedNodes()
		} else {
			ix[key] = n
			ix = n.getNestedNodes()
		}
	}
}

// is (foo|bar) is equal (bar|foo) ?
// (fo|f)(o|oo)

type group struct {
	key string
	end bool
	value [][]node
	expressions []string
	nested index
}

func (n *group) getKey() string {
	return n.key
}

func (n *group) getNestedNodes() index {
	return n.nested
}

func (n *group) getExpressions() []string {
	return n.expressions
}

func (n *group) isEnd() bool {
	return n.end
}

type namedGroup struct {
	key string
	name string
	value [][]node
	end bool
	expressions []string
	nested index
}

func (n *namedGroup) getKey() string {
	return n.key
}

func (n *namedGroup) getNestedNodes() index {
	return n.nested
}

func (n *namedGroup) getExpressions() []string {
	return n.expressions
}

func (n *namedGroup) isEnd() bool {
	return n.end
}

type notCapturedGroup struct {
	key string
	value [][]node
	end bool
	expressions []string
	nested index
}

func (n *notCapturedGroup) getKey() string {
	return n.key
}

func (n *notCapturedGroup) getNestedNodes() index {
	return n.nested
}

func (n *notCapturedGroup) getExpressions() []string {
	return n.expressions
}

func (n *notCapturedGroup) isEnd() bool {
	return n.end
}

type char struct {
	key string
	value rune
	end bool
	expressions []string
	nested index
}

func (n *char) getKey() string {
	return n.key
}

func (n *char) getNestedNodes() index {
	return n.nested
}

func (n *char) getExpressions() []string {
	return n.expressions
}

func (n *char) isEnd() bool {
	return n.end
}

type dot struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *dot) getKey() string {
	return n.key
}

func (n *dot) getNestedNodes() index {
	return n.nested
}

func (n *dot) getExpressions() []string {
	return n.expressions
}

func (n *dot) isEnd() bool {
	return n.end
}

type digit struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *digit) getKey() string {
	return n.key
}

func (n *digit) getNestedNodes() index {
	return n.nested
}

func (n *digit) getExpressions() []string {
	return n.expressions
}

func (n *digit) isEnd() bool {
	return n.end
}

type nonDigit struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *nonDigit) getKey() string {
	return n.key
}

func (n *nonDigit) getNestedNodes() index {
	return n.nested
}

func (n *nonDigit) getExpressions() []string {
	return n.expressions
}

func (n *nonDigit) isEnd() bool {
	return n.end
}

type word struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *word) getKey() string {
	return n.key
}

func (n *word) getNestedNodes() index {
	return n.nested
}

func (n *word) getExpressions() []string {
	return n.expressions
}

func (n *word) isEnd() bool {
	return n.end
}

type nonWord struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *nonWord) getKey() string {
	return n.key
}

func (n *nonWord) getNestedNodes() index {
	return n.nested
}

func (n *nonWord) getExpressions() []string {
	return n.expressions
}

func (n *nonWord) isEnd() bool {
	return n.end
}

type space struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *space) getKey() string {
	return n.key
}

func (n *space) getNestedNodes() index {
	return n.nested
}

func (n *space) getExpressions() []string {
	return n.expressions
}

func (n *space) isEnd() bool {
	return n.end
}

type nonSpace struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *nonSpace) getKey() string {
	return n.key
}

func (n *nonSpace) getNestedNodes() index {
	return n.nested
}

func (n *nonSpace) getExpressions() []string {
	return n.expressions
}

func (n *nonSpace) isEnd() bool {
	return n.end
}

type startOfLine struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *startOfLine) getKey() string {
	return n.key
}

func (n *startOfLine) getNestedNodes() index {
	return n.nested
}

func (n *startOfLine) getExpressions() []string {
	return n.expressions
}

func (n *startOfLine) isEnd() bool {
	return n.end
}

type endOfLine struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *endOfLine) getKey() string {
	return n.key
}

func (n *endOfLine) getNestedNodes() index {
	return n.nested
}

func (n *endOfLine) getExpressions() []string {
	return n.expressions
}

func (n *endOfLine) isEnd() bool {
	return n.end
}

type startOfString struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *startOfString) getKey() string {
	return n.key
}

func (n *startOfString) getNestedNodes() index {
	return n.nested
}

func (n *startOfString) getExpressions() []string {
	return n.expressions
}

func (n *startOfString) isEnd() bool {
	return n.end
}

type endOfString struct {
	key string
	end bool
	expressions []string
	nested index
}

func (n *endOfString) getKey() string {
	return n.key
}

func (n *endOfString) getNestedNodes() index {
	return n.nested
}

func (n *endOfString) getExpressions() []string {
	return n.expressions
}

func (n *endOfString) isEnd() bool {
	return n.end
}

type rangeNode struct {
	key string
	from rune
	to rune
	nested index
	expressions []string
	end bool
}

func (n *rangeNode) getKey() string {
	return n.key
}

func (n *rangeNode) getNestedNodes() index {
	return n.nested
}

func (n *rangeNode) getExpressions() []string {
	return n.expressions
}

func (n *rangeNode) isEnd() bool {
	return n.end
}

type quantifier struct {
	key string
	from int
	to *int
	more bool
	value node
	end bool
	expressions []string
	nested index
}

func (n *quantifier) getKey() string {
	return n.key
}

func (n *quantifier) getNestedNodes() index {
	return n.nested
}

func (n *quantifier) getExpressions() []string {
	return n.expressions
}

func (n *quantifier) isEnd() bool {
	return n.end
}

type positiveSet struct {
	key string
	value []node
	end bool
	expressions []string
	nested index
}

func (n *positiveSet) getKey() string {
	return n.key
}

func (n *positiveSet) getNestedNodes() index {
	return n.nested
}

func (n *positiveSet) getExpressions() []string {
	return n.expressions
}

func (n *positiveSet) isEnd() bool {
	return n.end
}

type negativeSet struct {
	key string
	value []node
	end bool
	expressions []string
	nested index
}

func (n *negativeSet) getKey() string {
	return n.key
}

func (n *negativeSet) getNestedNodes() index {
	return n.nested
}

func (n *negativeSet) getExpressions() []string {
	return n.expressions
}

func (n *negativeSet) isEnd() bool {
	return n.end
}

type buffer struct {
	data         []rune
	position     int

	// data string
	// positions []int (stack of last positions)
}

// Read - read next item, if greedy buffer keep position after reading.
func (b *buffer) Read(greedy bool) (rune, error) {
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
func (b *buffer) Seek(x int) {
	b.position = x
}

// Position - return current buffer position
func (b *buffer) Position() int {
	return b.position
}

// IsEOF - true if buffer ended.
func (b *buffer) IsEOF() bool {
	return b.position >= len(b.data)
}

// newBuffer - make buffer which can read text on input and use
// struct for positions.
func newBuffer(data []rune) c.Buffer[rune, int] {
	b := new(buffer)
	b.data = data
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

var (
	defaultParser = parseRegexp()
	none = struct{}{}

	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

func New(exps ...string) (Trie, error) {
	t := new(trie)
	t.nodes = make(index, len(exps))
	t.Add(exps...)
	return t, nil
}

func parseRegexp(except ...rune) expressionParser {
	var (
		regexp expressionParser
		groups parser
	)

	if len(except) == 0 {
		groups = choice(
			parseNotCapturedGroup(regexp),
			parseNamedGroup(regexp),
			parseGroup(regexp),
		)
	} else {
		nestedRegexp := parseRegexp(append(except, ')', '|')...)

		groups = choice(
			parseNotCapturedGroup(nestedRegexp),
			parseNamedGroup(nestedRegexp),
			parseGroup(nestedRegexp),
		)
	}

	characters := choice(
		parseInvalidQuantifier(),
		parseEscapedMetacharacters(),
		parseDot(),
		parseDigit(),
		parseNonDigit(),
		parseWord(),
		parseNonWord(),
		parseSpace(),
		parseNonSpace(),
		parseStartOfLine(),
		parseEndOfLine(),
		parseStartOfString(),
		parseEndOfString(),
		parseCharacter(except...),
	)

	setsCombinatrors := choice( // where dot?
		parseRange(append(except, ']')...),
		parseEscapedMetacharacters(),
		parseDigit(),
		parseNonDigit(),
		parseWord(),
		parseNonWord(),
		parseSpace(),
		parseNonSpace(),
		parseStartOfLine(),
		parseEndOfLine(),
		parseStartOfString(),
		parseEndOfString(),
		parseCharacter(except...),
	)

	sets := choice(
		parsePositiveSet(setsCombinatrors),
		parseNegativeSet(setsCombinatrors),
	)

	parse := c.Some(
		0,
		parseOptionalQuantifier(
			choice(
				sets,
				groups,
				characters,
			),
		),
	)

	return parse
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
		parsers[i] = parseEscapedMetacharacter(c)
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

		buf.Position()

		x := char{
			value: value,
			end: buf.IsEOF(),

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

					_, err := skip(buf)
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

					return q,  nil
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
		q.end = buf.IsEOF()

		return &q, nil
	}
}

func SkipString(data string) c.Combinator[rune, int, struct{}] {
	return func(buffer c.Buffer[rune, int]) (struct{}, error) {
		l := len(data)
		for _, x := range data {
			r, err := buffer.Read(true)
			if err != nil {
				return none, err
			}
			if x != r {
				return none, c.NothingMatched
			}
			l =- 1
		}

		if l != 0 {
			return none, c.NothingMatched
		}

		return none, nil
	}
}

func parseCharacter(except ...rune) parser {
	char := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := char(buf)
		if err != nil {
			return nil, err
		}

		x := dot{
			end: buf.IsEOF(),
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
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseDigit() parser {
	parse := SkipString("\\d")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := digit{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseNonDigit() parser {
	parse := SkipString("\\D")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonDigit{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseWord() parser {
	parse := SkipString("\\w")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := word{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseNonWord() parser {
	parse := SkipString("\\w")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonWord{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseSpace() parser {
	parse := SkipString("\\s")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := space{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseNonSpace() parser {
	parse := SkipString("\\S")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonSpace{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseStartOfLine() parser {
	parse := SkipString("\\^")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := startOfLine{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseEndOfLine() parser {
	parse := SkipString("\\$")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := endOfLine{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseStartOfString() parser {
	parse := SkipString("\\A")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := startOfString{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseEndOfString() parser {
	parse := SkipString("\\z")

	return func(buf c.Buffer[rune, int]) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := endOfString{
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseGroup(expression expressionParser) parser {
	sep := c.Eq[rune, int]('|')
	union := parens(c.SepBy1(0, expression, sep))

	return func(buf c.Buffer[rune, int]) (node, error) {
		variants, err := union(buf)
		if err != nil {
			return nil, err
		}

		x := group{
			value: variants,
			end: buf.IsEOF(),
		}

		return &x, nil
	}
}

func parseNotCapturedGroup(expression expressionParser) parser {
	sep := c.Eq[rune, int]('|')
	union := c.SepBy1[rune, int](0, expression, sep)
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
				value: variants,
				end: buf.IsEOF(),
			}

			return &x, nil
		},
	)
}

func parseNamedGroup(expression expressionParser, except ...rune) parser {
	sep := c.Eq[rune, int]('|')
	union := c.SepBy1[rune, int](1, expression, sep)
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
				name: string(name),
				value: variants,
				end: buf.IsEOF(),
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
			value: set,
			end: buf.IsEOF(),
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
			value: set,
			end: buf.IsEOF(),
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
			from: f,
			to: t,
		}

		return &x, nil
	}
}
