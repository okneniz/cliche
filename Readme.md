# Cliche

![Downloads](https://img.shields.io/github/downloads/okneniz/cliche/total)
![Contributors](https://img.shields.io/github/contributors/okneniz/cliche?color=dark-green)
![Forks](https://img.shields.io/github/forks/okneniz/cliche?style=social)
![Stargazers](https://img.shields.io/github/stars/okneniz/cliche?style=social)
![Issues](https://img.shields.io/github/issues/okneniz/cliche)
![License](https://img.shields.io/github/license/okneniz/cliche) 

Regular expressions engine for batch processing.

Main features:

- store expressions in trie like data structure and match few expression by one comparison
- bring several expressions into a unified form
- custom expression parsing

## Trie like data structure

Cliche compile expressions to chain of nodes and than add this chain to tree.
Every node have they own key.
When adding a new chain to the tree and finding the same key,
a new node isn't created, the new expression is simply added to it.
This way, the tree tries to be as minimally branched as possible,
which is beneficial when scanning text.

## Compaction / Unification

Cliche unify expression by few methods.

Character classes stored as [range table](https://pkg.go.dev/unicode#RangeTable).
All expression bellow the same and have them same one node in tree:
- [a-z1-2]
- [1-2a-z]
- [12a-z]
- [1a-z2]
- [1-2[a-z]]
- [[1-2][a-z]]
- [12[a-z]]
- [12a[b-z]]

Single characted stored as character class too.

Quantificators unified tooЖ

- `x+` equal `x{1,}`
- `x*` equal `x{0,}`
- `x?` equal `x{0,1}` and `x{,1}`

Comments removed in simple cases.

For example `x` equal `(?#123)x` and stored the same.

This way scanning few explessions sometimes is equal to scan one.

## Installation

```bash
go get github.com/okneniz/cliche
```

## Quick start

```
package main

import (
	"fmt"

	"github.com/okneniz/cliche"
)

func main() {
	tree := cliche.New(cliche.DefaultParser)

	tree.Add(
		"a[0123-9]+",
		"a[0123-9]{1,}",
	)

	fmt.Println("tree:")
	fmt.Println(tree.String())

	text := "Text with a1, b, c32."
	fmt.Println("scan text:", text)

	for _, match := range tree.Match(text) {
		fmt.Printf("text: %s\n", match.SubString())
		fmt.Printf("bounds: %v\n", match.Span())
		fmt.Println("regexps:")
		for _, regexp := range match.Expressions() {
			fmt.Printf("\t%v\n", regexp)
		}
	}
}
```

Output:

```
tree:
[
 {
  "key": "[97]",
  "type": "*node.class",
  "nested": [
   {
    "key": "[R16(48-57)]+",
    "type": "*node.quantifier",
    "expressions": [
     "a[0123-9]+",
     "a[0123-9]{1,}"
    ],
    "value": {
     "key": "[R16(48-57)]",
     "type": "*node.class"
    }
   }
  ]
 }
]

scan text:
match 0
	text: Text with a1, b, c32.
	bounds: [10-11]
	regexps: [a[0123-9]+ a[0123-9]{1,}]
```

## Documentation

[GoDoc documentation](https://pkg.go.dev/github.com/okneniz/cliche).

## Compabilities

Cliche have default compabilities common for most regular expressions engine.
You can configure your own or copy behaviour of exists engine. 

### Basic syntax

- `|` alternation
- `(...)` parentheses `()` group parts of a regular expression, allowing you to apply quantifiers or other operations to the group as a whole.
- `[...]` character class
- `\` escape (enable or disable meta character)
- postfix expressions as quantifiers

### Predefined engines

- [Onigmo](https://github.com/okneniz/cliche/tree/master/onigmo)
- RE2

## Roadmap

See the [open issues](https://github.com/okneniz/cliche/issues) for a list of proposed features (and known issues).

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.
* If you have suggestions for adding or removing projects, feel free to [open an issue](https://github.com/okneniz/cliche/issues/new) to discuss it, or directly create a pull request after you edit the *README.md* file with necessary changes.
* Please make sure you check your spelling and grammar.
* Create individual PR for each suggestion.

### Creating A Pull Request

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request
