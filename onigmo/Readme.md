# Onigmo engine

Default [Ruby](https://github.com/k-takata/Onigmo) regular expression engine.

### Predefined characters

| support |characters| description | UTF-8 code |
|--|--|--|--|
|✅| `\t` | horizontal tab | `0x09` |
|✅| `\v` | vertical tab | `0x0B` |
|✅| `\n` | newline (line feed) | `0x0A` |
|✅| `\r` | carriage return | `0x0D` |
|✅| `\b` | backspace | `0x08` |
|✅| `\f` | form feed | `0x0C` |
|✅| `\a` | bell | `0x07` |
|✅| `\e` | escape | `0x1B` |

### Different characters defenitions

| support |characters| description | UTF-8 code |
|--|--|--|--|
|❌| `\nnn` | octal char | encoded byte value |
|✅| `\xHH` | hexadecimal char | encoded byte value |
|❌| `\x{7HHHHHHH}` | wide hexadecimal char | character code point value |
|✅| `\uHHHH` | wide hexadecimal char| character code point value |
|❌| `\cx` | control char | character code point value |
|❌| `\C-x` | control char | character code point value |
|❌| `\M-x` | meta char | character code point value |
|❌| `\M-\C-x` | meta control char| character code point value |
|✅| `\o{nnn}` | octal char | character code point value

 * `\b` as backspace is effective in character class only

### Predefined meta characters

| support | characters | match |
|--|--|--|
|✅| `.` | any character (except newline) |
|✅| `\w` | any word character (letter, number, underscore) |
|✅| `\W` | any non-word character |
|✅| `\s` | any whitespace character |
|✅| `\S` | any non-whitespace character |
|✅| `\d` | any decimal digit character |
|✅| `\D` | any non-decimal digit character |
|✅| `\h` | any hexadecimal-digit char `[0-9a-fA-F]` |
|✅| `\H` | any non-hexadecimal-digit char |


### Character Property

| support | characters | match |
|--|--|--|
|✅|`\p{property-name}`| match character with [property](https://pkg.go.dev/unicode#pkg-variables) |
|✅|`\P{property-name}`| match character without [property](https://pkg.go.dev/unicode#pkg-variables)|
|✅|`\p{^property-name}`| match character without [property](https://pkg.go.dev/unicode#pkg-variables) |


### Quantifiers

#### Greedy

| support | characters | match |
|--|--|--|
|✅| `?` | 1 or 0 times |
|✅| `*` | 0 or more times |
|✅| `+` | 1 or more times |
|✅| `{n,m}` | at least n but no more than m times |
|✅| `{n,}` | at least n times |
|✅| `{,n}` | at least 0 but no more than n times ({0,n}) |
|✅| `{n}` | n times |

#### Reluctant

| support | characters | match |
|--|--|--|
|❌| `??` | 1 or 0 times |
|❌| `*?` | 0 or more times |
|❌| `+?` | 1 or more times |
|❌| `{n,m}?` | at least n but not more than m times |
|❌| `{n,}?` | at least n times |
|❌| `{,n}?` | at least 0 but not more than n times (== {0,n}?) |

#### Possessive 

Possesive - greedy and does not backtrack once match.

| support | characters | match |
|--|--|--|
|❌| `?+` | 1 or 0 times |
|❌| `*+` | 0 or more times |
|❌| `++` | 1 or more times |

### Anchors

| support | characters | match |
|--|--|--|
|✅| `^` | beginning of the line |
|✅| `$` | end of the line |
|✅| `\b` | word boundary |
|✅| `\B` | non-word boundary |
|✅| `\A` | beginning of string |
|✅| `\Z` | end of string, or before newline at the end |
|✅| `\z` | end of string |
|❌| `\G` | where the current search attempt begins |

### Character classes

| support | syntax | match |
|--|--|--|
|✅| `^...` | negative class (lowest precedence) |
|✅| `x-y` | range from x to y |
|✅| `[...]` | set (character class in character class) |
|❌| `..&&..` | intersection (low precedence, only higher than ^) ex. [a-w&&[^c-g]z] ==> ([a-w] AND ([^c-g] OR z)) ==> [abh-w] |

### Bracket ([:xxxxx:], negate [:^xxxxx:])

| support | construction | match |
|--|--|--|
|✅| `[:alnum:]` | `Letter \| Mark \| Decimal_Number` |
|✅| `[:alpha:]` | `Letter \| Mark` |
|✅| `[:ascii:]` | `0000 - 007F` |
|✅| `[:blank:]` | `Space_Separator \| 0009` |
|✅| `[:cntrl:]` | `Control \| Format \| Unassigned \| Private_Use \| Surrogate` |
|✅| `[:digit:]` | `Decimal_Number` |
|✅| `[:graph:]` | `[[:^space:]] && ^Control && ^Unassigned && ^Surrogate` |
|✅| `[:lower:]` | `Lowercase_Letter` |
|✅| `[:print:]` | `[[:graph:]] \| Space_Separator` |
|✅| `[:punct:]` | `Connector_Punctuation \| Dash_Punctuation \| Close_Punctuation \| Final_Punctuation \| Initial_Punctuation \| Other_Punctuation \| Open_Punctuation \| 0024 \| 002B \| 003C \| 003D \| 003E \| 005E \| 0060 \| 007C \| 007E` \|
|✅| `[:space:]` | `Space_Separator \| Line_Separator \| Paragraph_Separator \| 0009 \| 000A \| 000B \| 000C \| 000D \| 0085` |
|✅| `[:upper:]` | `Uppercase_Letter` |
|✅| `[:xdigit:]` | `0030 - 0039 \| 0041 - 0046 \| 0061 - 0066 \| (0-9, a-f, A-F)` |
|✅| `[:word:]` | `Letter \| Mark \| Decimal_Number \| Connector_Punctuation` |

### Extended groups

| support | construction | match |
|--|--|--|
|✅| `(?:subexp)` | non-capturing group |
|✅| `(subexp)` | capturing group |
|✅| `(?<name>subexp)` | define named group |

### Backreferences

When we say "backreference a group," it actually means, "re-match the same text matched by the subexp in that group."

| support | construction | match |
|--|--|--|
|✅| `(exp)\1` | backrefernces by index |
|✅| `(?<name>exp)\k<name>` | backreferences by name |
|❌| `(?<name>exp)\g<name>` | call a group by name |

❌ backreference with recursion level

```
  \n  \k<n>     \k'n'     (n >= 1) backreference the nth group in the regexp
      \k<-n>    \k'-n'    (n >= 1) backreference the nth group counting
                          backwards from the referring position

  When backreferencing with a name that is assigned to more than one groups,
  the last group with the name is checked first, if not matched then the
  previous one with the name, and so on, until there is a match.

  * Backreference by number is forbidden if any named group is defined and
    ONIG_OPTION_CAPTURE_GROUP is not set.

  * ONIG_SYNTAX_PERL: \g{n}, \g{-n} and \g{name} can also be used.
    If a name is defined more than once in Perl syntax, only the left-most
    group is checked.

  backreference with recursion level

    (n >= 1, level >= 0)

    \k<n+level>  \k'n+level'
    \k<n-level>  \k'n-level'
    \k<-n+level> \k'-n+level'
    \k<-n-level> \k'-n-level'

    \k<name+level> \k'name+level'
    \k<name-level> \k'name-level'
```
### Subexp calls ("Tanaka Akira special")

```
  When we say "call a group," it actually means, "re-execute the subexp in
  that group."

  \g<0>     \g'0'     call the whole pattern recursively
  \g<n>     \g'n'     (n >= 1) call the nth group
  \g<-n>    \g'-n'    (n >= 1) call the nth group counting backwards from
                      the calling position
  \g<+n>    \g'+n'    (n >= 1) call the nth group counting forwards from
                      the calling position
  \g<name>  \g'name'  call the group with the specified name

  * Left-most recursive calls are not allowed.

    ex. (?<name>a|\g<name>b)    => error
        (?<name>a|b\g<name>c)   => OK

  * Calls with a name that is assigned to more than one groups are not
    allowed in ONIG_SYNTAX_RUBY.

  * Call by number is forbidden if any named group is defined and
    ONIG_OPTION_CAPTURE_GROUP is not set.

  * The option status of the called group is always effective.

    ex. /(?-i:\g<name>)(?i:(?<name>a)){0}/.match("A")
```

### Assertions
 
https://www.regular-expressions.info/lookaround.html

The difference is that lookaround actually matches characters, but then gives up the match, 
returning only the result: match or no match. That is why they are called “assertions”.

| support | construction | match |
|--|--|--|
|✅| `(?=subexp)` | look-ahead |
|✅| `(?<=subexp)` | look-behind |
|✅| `(?!subexp)` | negative look-ahead |
|✅| `(?<!subexp)` | negative look-behind |
|✅| `(?>subexp)` | atomic group |
|❌| `(?~subexp)` | absence operator |
|✅| `prefix\Ksubexp` | another expression of look-behind. Keep the stuff left of the \K, don't include it in the result. |

```
Subexp of look-behind must be fixed-width.
But top-level alternatives can be of various lengths.
ex. (?<=a|bc) is OK. (?<=aaa(?:b|cd)) is not allowed.

In negative look-behind, capturing group isn't allowed,
but non-capturing group (?:) is allowed.

Atomic group no backtracks in subexp.

  (?~subexp)        absence operator (experimental)
                    Matches any string which doesn't contain any string which
                    matches subexp.
                    More precisely, (?~subexp) matches the complement set of
                    a set which .*subexp.* matches. This is regular in the
                    meaning of formal language theory.
                    Similar to (?:(?!subexp).)*, but easy to write.

                    E.g.:
                      (?~abc) matches: "", "ab", "aab", "ccdd", etc.
                      It doesn't match: "abc", "aabc", "ccabcdd", etc.

                      \/\*(?~\*\/)\*\/ matches C style comments:
                        "/**/", "/* foobar */", etc.

                      \A\/\*(?~\*\/)\*\/\z doesn't match "/**/ */".
                      This is different from \A\/\*.*?\*\/\z which uses a
                      reluctant quantifier (.*?).

                      Unlike (?:(?!abc).)*c, (?~abc)c matches "abc", because
                      (?~abc) matches "ab".

                      (?~) never matches.
```

### Condition

Matches yes-subexp if (cond) yields a true value, matches no-subexp otherwise.

https://www.regular-expressions.info/conditional.html

| support | construction | description |
|--|--|--|
|✅| `(?(cond)yes-subexp\|no-subexp)` | checks if the numbered capturing group has matched something (n >= 1) |
|✅| `(?(<cond>)yes-subexp\|no-subexp)` | checks if a group with the given name has matched something |
|✅| `(?(n)yes-subexp)` | condition with one branch (n >= 1)|
|✅| `(?(<cond>)yes-subexp)` | condition with one branch |


### Options

Options change default behaviour.

#### Scan options

Scan option passed as arguments to `Scan` method and change behavoir all expression in three.

| support | description |
|--|--|
|✅| ignore case |
|✅| multi-line (dot (.) also matches newline) | 

#### Expression options

| support | example | description |
|--|--|--|
|✅| `/(?i)A/` or `/(?i:A)/` match 'a' | ignore case |
|✅| `/(?m).A/` or `/(?m:.A)/` match '\na' | multi-line (dot (.) also matches newline) | 

#### Parser options

| support | example | description |
|--|--|--|
|❌| `/(?x)a\nb\c/` parsed as `/abc/` | extended form - expressions items separated by whitespaces |


### Other

| support | construction | description |
|--|--|--|
|✅| `(?# ... )` | comment |

### Corner cases

Usually linters hihglight them as mistakes.

| support | construction | description |
|--|--|--|
|❌|`//` | empty expression |
|❌|`()` | empty group |
|❌|`/a\|/`, `/\|a/` or `(a\|b\|)` | empty variant in alternation |
```

## Roadmap

https://www.regular-expressions.info/branchreset.html
https://www.regular-expressions.info/freespacing.html

- tests:
  - property-based testing
  - complex tests
  - split tests to different groups (maybe by tags), for example:
    - POSIX
    - ERE
    - BRE
    - RE2 / golang
    - Onigmo / ruby
    - V8 / JS
  - use testdata from another libs
    - RE2 / golang - https://github.com/golang/go/tree/master/src/regexp/testdata
    - Onigmo / ruby - https://github.com/k-takata/Onigmo/blob/master/test.rb
- matches list test
- add recursive calls \g
- think about reluctant and possessive quantifiers (Is it possible with this architecture?)

## Compaction

- alternation with one variant to this variant
- remove comments
- \w{3} -> \w\w\w
- \w{1} -> \w
- (?i)(?-i)(?i) -> (?i)
- (?im) -> (?i)(?m)
- (?i:foo) -> (?i)(?:foo)
? not captured group with alternation with one variant to this variant

- можно якоря / anchor типа \b жать как (?<!subexp)` / negative look-behind
- типа границы слова это что-то, что до слова и после слова (пробелы иoли знаки препинания)

## Validation

- empty class
- endless comments
- not fixed size lookaheads

## Features

для выражений типа [:printed:]

можно делать интересные штуки типа бинарной арифметики
- пользовтель добавляет свой [:something:]
- в его конфиге указывает match функцию, которая говорит ок или не ок

один из примеров xor на бинарное представление rune

- так пример можно будет расширить для других кодировок и range table-ов
- возможны регулярные выражения для бинарных данных

// custom:
// - brackets [[:cyrilic:]]
// - meta chars \ѣ
// - custom anything for binary data

// https://www.rfc-editor.org/rfc/rfc9485.html#name-implementing-i-regexp

// Добавить метод Nonsence который говорит какие выражения никогда не заматчатся
// - например из-за пустой range table
// - или из-за якоря \z после \A
// - или странного assertion / lookahead, например 1(?=3)2

// What is OnePass?
// https://github.com/golang/go/blob/master/src/regexp/onepass_test.go
//
// Is it optimization too? Or just compaction?

## Differences

// bnf / ebnf
//
// https://www2.cs.sfu.ca/~cameron/Teaching/384/99-3/regexp-plg.html
//
// https://swtch.com/~rsc/regexp/regexp2.html#posix
//
// https://www.rfc-editor.org/rfc/rfc9485.html#name-multi-character-escapes

// GROUPS name collision

// RUBY - /(?<first>\D)(?<first>\D)(?<first>\D)(?<fourth>\D)/.match('foobar').named_captures => {"first"=>"o", "fourth"=>"b"}
// JS - report about error

## Need check

// TODO: remove onMatch Callback params (required only for quantifier?)
// pass quantifier as scanner?
// or maybe just reduce calls of onMatch() (only if it's leaf?)

// JAVASCRIPT - /(a)\2/u; // SyntaxError: Invalid regular expression: Invalid escape

// проверить как парсятся ключи в alternation + groups
// внутри <([^<>]+)>[^<>]+(<(span|em|i|b)>([^<>]+)<\/\3>)[^<>]+<\/\1>
// находил <(b,s,e,i)>
// кажется баг где-то

// TODO : использовть bitset для ключей Node
// для utf нужен make([]byte, unicode.MaxRune / sozeOf(byte))
// если predicate сработал, то ставить bit = 1, иначе bit = 0

// проверять что группы с одинковыми именами но разным выражением - не жмутся
// проверять что альтернативы с одинаковыми вариантами жмутся в обычную ноду (на уровне ключа)

## Need tests

// add test for
// <([^<>]+)>[^<>]*(<(span|em|i|b)>([^<>]+)<\/\\3>)[^<>]*<\/\\1>

// add test for different captures for string "123$ 231€ 321₽"
//  (\\d+(?=(\\$|₽)))
//  (\\d+(?=\\$|₽))

// add test for "123$ 231€ 321₽"
// (((123)))

// add test for
// RUBY - /.{2}|abc|\s/.match("abc").to_a

// cost is /\$(?<=\$)10/ match "cost is $10"

// TODO : check this
// https://stackoverflow.com/questions/2973436/regex-lookahead-lookbehind-and-atomic-groups
// (?<=foo)bar(?=bar)    finds the 1st bar ("bar" with "foo" before it and "bar" after it)
// it doesn't work on rubular for sting "foobar"

// TODO : check this too

// (\d+(?!(\$|₽)))
// 123$ 231€ 321₽
// rubular render something strange - https://rubular.com
// problem copy euro symbol to irb

// TODO : add more tests for back references
// with all kind of groups

##  Property tests 

- size нод не должно меняться при изменениях дерева
- элементы для encoding/unicode.NewTableFor всегда дают одинковый результата

// https://www.regular-expressions.info/lookaround.html 
// Regex Engine Internals
// - The regex q(?=u)i can never match anything. 

// Добавить тесты на утверждения типа "никогд не заматчится"
// https://www.regular-expressions.info/lookaround.html
//
// Lookaround Is Atomic
//
// For this reason, the regex (?=(\d+))\w+\1 never matches 123x12.
//
// But the regex (?=(\d+))\w+\1 does match 56x56 in 456x56

// ((?<=(^))(.+)(?=($)))
//
// must match any string

// TODO : try to explain it in doc for contributors
//
// split to another groups of options
//
// common:
//   - chars (value)
//     - as is (a, b, 1)
//     - with prefix (\u{123}, \x017)
//   - escaped meta chars (range of value)
//   - classes (range of value)
//
// not in class:
//   - groups
//   - assertions (lookahead / lookbehind)
//   - alternative
//   - anchors: (match positions)
//   	- ^, $
//   	- \A, \z
// 	 - spec symbols - [(|
//   - quantifiers *+?
//   - meta chars - ^$.
//
// in class:
//   - bracket
//   - ranges from chars
// 	 - spec symbols - ^])

// node
//   - match node (chars, classes)
//     - return false if bounds out of ranges)
//	   - yield only not empty span
//   - position node (anchor)
//	   - yield only empty span
//   - special consuctions (group, alternation, assertions)
//     - capture internal sub expression

