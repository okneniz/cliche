# Cliche

Regular expressions engine.

Main features:
- store expressions in compacted form - as trie like data structure
- with focus on efficient matching - can match / unmatch few expressions by one comparison
- custom parsing

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
