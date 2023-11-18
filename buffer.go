package regular

import (
	"fmt"

	c "github.com/okneniz/parsec/common"
)

type simpleBuffer struct {
	data     []rune
	position int

	// data string
	// positions []int (stack of last positions)
}

type TextBuffer interface {
	ReadAt(int) rune
	Size() int
	Substring(int, int) (string, error)
	String() string
}

var _ TextBuffer = &simpleBuffer{}
var _ c.Buffer[rune, int] = &simpleBuffer{}

// newBuffer - make buffer which can read text on input
func newBuffer(str string) *simpleBuffer {
	b := new(simpleBuffer)
	b.data = []rune(str)
	b.position = 0
	return b
}

// Read - read next item, if greedy buffer keep position after reading.
func (b *simpleBuffer) Read(greedy bool) (rune, error) {
	if b.IsEOF() {
		return 0, c.EndOfFile
	}

	x := b.data[b.position]

	if greedy {
		b.position++
	}

	return x, nil
}

func (b *simpleBuffer) ReadAt(idx int) rune {
	return b.data[idx]
}

func (b *simpleBuffer) Size() int { // TODO : check for another runes
	return len(b.data)
}

func (b *simpleBuffer) String() string {
	return fmt.Sprintf("Buffer(%s, %d)", string(b.data), b.position)
}

func (b *simpleBuffer) Substring(from, to int) (string, error) {
	if from > to {
		// TODO : use OutOfBounds?
		return "", fmt.Errorf(
			"invalid bounds for substring: from=%d to=%d size=%d",
			from,
			to,
			len(b.data),
		)
	}

	if from < 0 || from >= len(b.data) || to >= len(b.data) {
		return "", fmt.Errorf(
			"out of bounds buffer: from=%d to=%d size=%d",
			from,
			to,
			len(b.data),
		)
	}

	return string(b.data[from : to+1]), nil
}

// Seek - change buffer position
func (b *simpleBuffer) Seek(x int) {
	b.position = x
}

// Position - return current buffer position
func (b *simpleBuffer) Position() int {
	return b.position
}

// IsEOF - true if buffer ended
func (b *simpleBuffer) IsEOF() bool {
	return b.position >= len(b.data)
}
