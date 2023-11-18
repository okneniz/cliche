package regular

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// https://www.regular-expressions.info/repeat.html

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
//
// https://www.rfc-editor.org/rfc/rfc9485.html#name-multi-character-escapes

type Trie interface {
	Add(...string) error
	Size() int
	MarshalJSON() ([]byte, error)
	String() string
	Match(string) []*FullMatch
}

var _ Trie = new(trie)

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
			x.addExpression(str)
		}
	})

	key := newNode.getKey()

	if prev, exists := t.nodes[key]; exists {
		prev.merge(newNode)
	} else {
		t.nodes[key] = newNode
	}
}

func (t *trie) Match(text string) []*FullMatch {
	if len(text) == 0 {
		return nil
	}

	input := newBuffer(text)
	matches := newMatchesList[match]()
	groups := newCaptures() // TODO : use node as key for unnamed groups to avoid generate string ID
	namedGroups := newCaptures()

	// DS for best matches - https://web.engr.oregonstate.edu/~erwig/diet/
	acc := make(map[node]*matchesList[*FullMatch])

	var scanner *fullScanner

	scanner = newFullScanner(
		groups,
		namedGroups,
		func(n node, from, to int, empty bool) {
			matches.push(
				match{
					from:  from,
					to:    to,
					node:  n,
					empty: empty,
				},
			)

			begin := scanner.FirstMatch()
			end := scanner.LastMatch()

			beginSubstring := scanner.FirstNotEmptyMatch()
			endSubstring := scanner.LastNotEmptyMatch()

			fmt.Println("scanner", scanner)

			m := &FullMatch{
				expressions: n.getExpressions().toSlice(),
				from:        begin.From(),
				to:          end.To(),
				groups:      groups.ToSlice(),
				namedGroups: namedGroups.ToMap(),
			}

			if m.from >= input.Size() {
				m.from = input.Size() - 1
			}

			if m.to >= input.Size() {
				m.to = input.Size() - 1
			}

			if beginSubstring != nil && endSubstring != nil {
				subString, err := input.Substring(
					beginSubstring.From(),
					endSubstring.To(),
				)

				if err != nil {
					// TODO : how to handle error
					fmt.Println("error", err)
				}

				m.subString = subString
			} else {
				m.empty = true
			}

			fmt.Printf("full match: %v\n", m)

			if list, exists := acc[n]; exists {
				list.push(m)
			} else {
				newList := newMatchesList[*FullMatch]()
				newList.push(m)
				acc[n] = newList
			}

			fmt.Println(" ")
		},
	)

	from := 0
	to := input.Size() - 1

	// - как правильно матчить
	// - как избегать лишних сканирований?

	for _, n := range t.nodes {
		nextFrom := from

		for nextFrom <= to {
			n.match(scanner, input, nextFrom, to, func(x node, f, t int, _ bool) {
				// if n.isEnd() {
				// 	fmt.Printf("match %v '%s' from %d to %d\n", x.getExpressions(), x.getKey(), nextFrom, nextTo)
				// }
			})

			longestMatch := matches.maximum() // maybe rename to best?

			if longestMatch != nil {
				nextFrom = longestMatch.To() + 1
			} else {
				nextFrom += 1
			}

			scanner.Rewind(0)
			matches.clear()
		}
	}

	result := make([]*FullMatch, 0, len(acc))
	for _, list := range acc {
		for _, item := range list.toMap() {
			result = append(result, item)
		}
	}

	return result
}
