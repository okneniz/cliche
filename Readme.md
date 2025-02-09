# Cliche

Regular expressions engine.

Main features:
- store expressions in compacted form - as trie like data structure
- with focus on efficient matching - can match / unmatch few expressions by one comparison
- custom parsing

## Compabilities

### Basic syntax

  - `\` escape (enable or disable meta character)
  - `|` alternation
  - `(...)` group
  - `[...]` character class

### Predefined characters

| support |characters| description | UTF-8 code |
|--|--|--|--|
|❌| `\t` | horizontal tab | `0x09` |
|❌| `\v` | vertical tab | `0x0B` |
|✅| `\n` | newline (line feed) | `0x0A` |
|❌| `\r` | carriage return | `0x0D` |
|❌| `\b` | backspace | `0x08` |
|❌| `\f` | form feed | `0x0C` |
|❌| `\a` | bell | `0x07` |
|❌| `\e` | escape | `0x1B` |

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
|❌|`\p{^property-name}`| match character without [property](https://pkg.go.dev/unicode#pkg-variables) |


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
|❌| `\b` | word boundary |
|❌| `\B` | non-word boundary |
|✅| `\A` | beginning of string |
|❌| `\Z` | end of string, or before newline at the end |
|✅| `\z` | end of string |
|❌| `\G` | where the current search attempt begins |

### Character classes

| support | syntax | match |
|--|--|--|
|✅| `^...` | negative class (lowest precedence) |
|✅| `x-y` | range from x to y |
|❌| `[...]` | set (character class in character class) |
|❌| `..&&..` | intersection (low precedence, only higher than ^) ex. [a-w&&[^c-g]z] ==> ([a-w] AND ([^c-g] OR z)) ==> [abh-w] |


### Bracket ([:xxxxx:], negate [:^xxxxx:])

| support | characters | match |
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

| support | characters | match |
|--|--|--|
|✅| `(?:subexp)` | non-capturing group |
|✅| `(subexp)` | capturing group |
|✅| `(?<name>subexp)` | define named group |

### Backreferences

When we say "backreference a group," it actually means, "re-match the same text matched by the subexp in that group."

| support | characters | match |
|--|--|--|
|✅| `(exp)\1` | backrefernces by index |
|✅| `(?<name>exp)\k<name>` | backreferences by name |


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

### Asertions
 
https://www.regular-expressions.info/lookaround.html

The difference is that lookaround actually matches characters, but then gives up the match, 
returning only the result: match or no match. That is why they are called “assertions”.

| support | characters | match |
|--|--|--|
|✅| `(?=subexp)` | look-ahead |
|✅| `(?<=subexp)` | look-behind |
|✅| `(?!subexp)` | negative look-ahead |
|✅| `(?<!subexp)` | negative look-behind |
|❌| `(?>subexp)` | atomic group |
|❌| `(?~subexp)` | absence operator |

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

  \K  keep
      Another expression of look-behind. Keep the stuff left
      of the \K, don't include it in the result.

  Theoretical backgrounds are discussed in Tanaka Akira's
  paper and slide (both Japanese):

    * Absent Operator for Regular Expression
      https://staff.aist.go.jp/tanaka-akira/pub/prosym49-akr-paper.pdf
    * 正規表現における非包含オペレータの提案
      https://staff.aist.go.jp/tanaka-akira/pub/prosym49-akr-presen.pdf
```

### Condition

```
  (?(cond)yes-subexp), (?(cond)yes-subexp|no-subexp)
    conditional expression
    Matches yes-subexp if (cond) yields a true value, matches
    no-subexp otherwise.
    Following (cond) can be used:

    (n)  (n >= 1)
        Checks if the numbered capturing group has matched
        something.

    (<name>), ('name')
        Checks if a group with the given name has matched
        something.

        BUG: If the name is defined more than once, the
        left-most group is checked, but it should be the
        same as \k<name>.
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

  * ONIG_SYNTAX_PERL:
    Use (?&name), (?n), (?-n), (?+n), (?R) or (?0) instead of \g<>.
    Calls with a name that is assigned to more than one groups are allowed,
    and the left-most subexp is used.
```

### Options

```
  (?imxdau-imx)      option on/off
                         i: ignore case
                         m: multi-line (dot (.) also matches newline)
                         x: extended form

                       character set option (character range option)
                         d: Default (compatible with Ruby 1.9.3)
                            \w, \d and \s doesn't match non-ASCII characters.
                            \b, \B and POSIX brackets use the each encoding's
                            rules.
                         a: ASCII
                            ONIG_OPTION_ASCII_RANGE option is turned on.
                            \w, \d, \s and POSIX brackets doesn't match
                            non-ASCII characters.
                            \b and \B use the ASCII rules.
                         u: Unicode
                            ONIG_OPTION_ASCII_RANGE option is turned off.
                            \w (\W), \d (\D), \s (\S), \b (\B) and POSIX
                            brackets use the each encoding's rules.

  (?imxdau-imx:subexp) option on/off for subexp

  Behavior of an unnamed group (...) changes with the following conditions.
  (But named group is not changed.)

  case 1. /.../     (named group is not used, no option)

     (...) is treated as a capturing group.

  case 2. /.../g    (named group is not used, 'g' option)

     (...) is treated as a non-capturing group (?:...).

  case 3. /..(?<name>..)../   (named group is used, no option)

     (...) is treated as a non-capturing group.
     numbered-backref/call is not allowed.

  case 4. /..(?<name>..)../G  (named group is used, 'G' option)

     (...) is treated as a capturing group.
     numbered-backref/call is allowed.

  ('g' and 'G' options are argued in ruby-dev ML)

A-2. Original extensions

   + named group                  (?<name>...), (?'name'...)
   + named backref                \k<name>
   + subexp call                  \g<name>, \g<group-num>
```

### Other

```
(?#...)            comment
```

## Roadmap

- add atomic groups
- add keep \K
- add recursive calls \g
- add conditions
- add comments (?#...)
- add options
  - case insensetive
  - multi line
  - named groups
- refactor traverse
- split tests to different groups (maybe by tags), for example:
  - POSIX
  - ERE
  - BRE
  - RE2 / golang
  - Ruby / onigma
  - V8 / JS?
- use testdata from another libs
  - RE2 / golang - https://github.com/golang/go/tree/master/src/regexp/testdata
  - Onigmo / ruby - https://github.com/k-takata/Onigmo/blob/master/test.rb
- more tests
  - complex tests
  - property-based testing
- more compactions
  - quatifiers to sequence \w{3} -> \w\w\w
  - anchors to assertions / look-behind / look-aheads
- think about Reluctant and Reluctant quantifiers (Is it possible with this architecture?)

## Octal character defenitions can conflict with quantifiers

- `\o{nnn}` - 

## Compact

- \w{3} -> \w\w\w

- можно якоря / anchor типа \b жать как (?<!subexp)` / negative look-behind
- типа границы слова это что-то, что до слова и после слова (пробелы или знки препинания)

// TODO : return error for invalid escaped chars like '\x' (check on rubular)

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
// - custom anything

// https://www.rfc-editor.org/rfc/rfc9485.html#name-implementing-i-regexp


// Добавить метод Nonsence который говорит какие выражения никогда не заматчатся
// - например из-за пустой range table
// - или из-за якоря \z после \A
// - или странного assertion / lookahead, например 1(?=3)2

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

// What about expressions optimizations?
// How to automate it?
// Is it posssible? https://www.regular-expressions.info/alternation.html

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

// TODO :
//
// is it possible to capture empty string?
//
// example:
//
// (^)foo($)

// what abour nested empty captures

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
// может так можно убрать интерфейс node.Alternation?

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

