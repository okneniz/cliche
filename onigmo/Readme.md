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
