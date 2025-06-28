package parser

import (
	"fmt"

	"github.com/okneniz/cliche/quantity"
)

type ParsingError struct {
	Scope       string
	Expectation string
	Position    quantity.Interface
}

func (err *ParsingError) Error() string {
	return "nothing matched"
}

type MultipleParsingError struct {
	Errors []ParsingError
}

func (err *MultipleParsingError) Error() string {
	return "nothing matched"
}

func Explain(err MultipleParsingError) string {
	return fmt.Sprintf("%#v", err)
}
