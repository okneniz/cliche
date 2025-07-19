package parser

import (
	"errors"
	"fmt"
	"strings"

	c "github.com/okneniz/parsec/common"
)

type (
	Error interface {
		Position() int
		NestedErrors() []Error
		Error() string
	}

	expectationError struct {
		expectation string
		position    int
		actual      error
	}

	parsingError struct {
		position     int
		nestedErrors []Error
	}
)

func Expected(expect string, pos int, actual error) Error {
	return &expectationError{
		expectation: expect,
		position:    pos,
		actual:      actual,
	}
}

func (err *expectationError) Position() int {
	return err.position
}

func (err *expectationError) Error() string {
	if err.actual == nil || isParsecStandardError(err.actual) {
		return fmt.Sprintf(
			"expected %s at position %d",
			err.expectation,
			err.position,
		)
	}

	return fmt.Sprintf(
		"expected %s at position %d, but actual %s",
		err.expectation,
		err.position,
		err.actual.Error(),
	)
}

func (err *expectationError) NestedErrors() []Error {
	return nil
}

func MergeErrors(errs ...Error) Error {
	merged := make([]Error, 0, len(errs))
	for _, x := range errs {
		merged = append(merged, x.NestedErrors()...)
	}

	return &parsingError{
		nestedErrors: merged,
	}
}

func (err *parsingError) Position() int {
	return err.position
}

func (err *parsingError) Error() string {
	messages := make([]string, len(err.nestedErrors))

	for i, x := range err.nestedErrors {
		messages[i] = x.Error()
	}

	return strings.Join(messages, ", ")
}

func (err *parsingError) NestedErrors() []Error {
	return err.nestedErrors
}

func Explain(err Error) string {
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
