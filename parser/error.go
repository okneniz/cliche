package parser

import (
	"errors"
	"fmt"
	"strings"

	c "github.com/okneniz/parsec/common"
)

type (
	// TODO : rename to ErrExpectation?

	ParsingError struct {
		Expectation string
		Position    int
		Actual      error
	}

	MultipleParsingError struct {
		Errors []*ParsingError
	}
)

func Expected(expect string, pos int, actual error) *MultipleParsingError {
	err := &ParsingError{
		Expectation: expect,
		Position:    pos,
		Actual:      actual,
	}

	return &MultipleParsingError{
		Errors: []*ParsingError{
			err,
		},
	}
}

func (err *ParsingError) Error() string {
	if err.Actual == nil || isParsecStandardError(err.Actual) {
		return fmt.Sprintf(
			"expected %s at position %d",
			err.Expectation,
			err.Position,
		)
	}

	return fmt.Sprintf(
		"expected %s at position %d, but actual %s",
		err.Expectation,
		err.Position,
		err.Actual.Error(),
	)
}

func MergeErrors(errs ...*MultipleParsingError) *MultipleParsingError {
	total := 0
	for _, x := range errs {
		total += len(x.Errors) // TODO : may be uniq?
	}

	merged := make([]*ParsingError, 0, total)
	for _, x := range errs {
		merged = append(merged, x.Errors...)
	}

	return &MultipleParsingError{
		Errors: merged,
	}
}

func (err *MultipleParsingError) Error() string {
	messages := make([]string, len(err.Errors))

	for i, x := range err.Errors {
		messages[i] = x.Error()
	}

	return strings.Join(messages, ", ")
}

func Explain(err *MultipleParsingError) string {
	return fmt.Sprintf("%#v", err)
}

// TODO : rename NothingMatched to ErrNothingMatched
func isParsecStandardError(err error) bool {
	return errors.Is(err, c.NothingMatched) ||
		errors.Is(err, c.EndOfFile) ||
		errors.Is(err, c.NotEnoughElements)
}

// нужно ловить всего несколько видов ошибок:
// 1. конец файла или строки
// 2. NothingMatched?
// 3. не сматчилось то, что предпологалось / ожидалось (ParsingError)
// 4. мерж нескольких ошибок (MultipleError)
// 5. ошибки валидации node (наверно можно сделать через ParsingError точно так же)
//    - return error for invalid escaped chars like '\x' (check on rubular)
//    - subexp of look-behind must be fixed-width.
//      but top-level alternatives can be of various lengths.
//      ex. (?<=a|bc) is OK. (?<=aaa(?:b|cd)) is not allowed.
//
// непонятно нужен ли вообще обобщенный интерфейс
// например Scope наверно всегда может возвращать типизированную ошибку
// часть данных в этой типизировнной ошибки можно считать заранее
// только позиция в строке разная
//
// на каких ошибках нужно прерывть парсинг и нужно ли это вообще
// нужно ли это поведение повторять как в других движках
//
// цель - исключить рефлексии и анализ ошибок

// сложность:
// - непонятен интерфейс обощенной ошибки
// - второй и третий пункт одинаковы?
//   например NothingMatch можно конвертить в ParsingError со спец знчениями
// - каких данных будет достаточно для красивого отображения ошибки
// - нужно ли пользователю указывать scope name или parser name
