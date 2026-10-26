# Cliche

![Downloads](https://img.shields.io/github/downloads/okneniz/cliche/total)
![Contributors](https://img.shields.io/github/contributors/okneniz/cliche?color=dark-green)
![Forks](https://img.shields.io/github/forks/okneniz/cliche?style=social)
![Stargazers](https://img.shields.io/github/stars/okneniz/cliche?style=social)
![Issues](https://img.shields.io/github/issues/okneniz/cliche)
![License](https://img.shields.io/github/license/okneniz/cliche) 

Regular expressions engine for batch processing.

Main features:

- store expressions as tree and match few of them by one comparison
- [unified several expressions to one form](#compaction--unification)
- [custom expression parsing](#parsing-and-predefined-engines)

## Installation

```bash
go get github.com/okneniz/cliche
```

## Quick start

```golang
package main

import (
	"fmt"

	"github.com/okneniz/cliche"
)

func main() {
	tree := cliche.New(cliche.DefaultParser)

	tree.Add(
		"a[0123-9]+",
		"a[01-5[67-9]]{1,}",
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
     "a[01-5[67-9]]{1,}"
    ],
    "value": {
     "key": "[R16(48-57)]",
     "type": "*node.class"
    }
   }
  ]
 }
]

scan text: Text with a1, b, c32.
text: a1
bounds: [10-11]
regexps:
	a[01-5[67-9]]{1,}
	a[0123-9]+
```

## Documentation

[GoDoc documentation](https://pkg.go.dev/github.com/okneniz/cliche).


## Compaction / Unification

Cliche unify expression by few methods.

Character classes stored as [range table](https://pkg.go.dev/unicode#RangeTable).
All expression bellow the same:
- `[a-z1-2]`
- `[1-2a-z]`
- `[12a-z]`
- `[1a-z2]`
- `[1-2[a-z]]`
- `[[1-2][a-z]]`
- `[12[a-z]]`
- `[12a[b-z]]`

Single character stored as character class too.
All expression bellow the same:

- `a`
- `[a]`
- `[aaaa]`
- `[a-a]`
- `[a-aa]`

Quantificators unified too:

- `x+` equal `x{1,}`
- `x*` equal `x{0,}`
- `x?` equal `x{0,1}` and `x{,1}`

Comments removed in simple cases.

For example `x` equal `(?#123)x` and stored the same.

Group options unified too.
All expression bellow the same:

- `(?i:y)` eqaul `(?i)(?:y)(?-i)`
- `(?i-m:test)` equal `(?i-m)(test)(?m-i)`

Non-unique variants within an alternation are removed from it.
All expression bellow the same:

- `(a|b|c)`
- `(a|b|c|c)`
- `(a|b|b|c)`
- `(a|a|b|c)`
- `(a|[a]|b|c)`
- `(a|[a]|[b]|[c])`

You can see more examples [here](https://github.com/okneniz/cliche/blob/master/compaction_test.go).

Result of unification - reusing one path or branch by more than one expressions.
Scanner can match multiple expressions at once.
Of course unification not change behaviour of tree or scanner results.

## Parsing and predefined engines

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
- [RE2](https://github.com/okneniz/cliche/tree/master/re2)

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
