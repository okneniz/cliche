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
