package cliche

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

type Input interface {
	ReadAt(int) rune
	Size() int
	String() string
}

var _ Input = &simpleBuffer{}
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

func (b *simpleBuffer) Size() int {
	return len(b.data)
}

func (b *simpleBuffer) String() string {
	return fmt.Sprintf("Buffer(%s, %d)", string(b.data), b.position)
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
