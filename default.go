package cliche

import "github.com/okneniz/cliche/parser"

var (
	DefaultOptions = OnigmoOptions
	DefaultParser  = parser.NewParser(DefaultOptions...)
)
