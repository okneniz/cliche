package buf

import (
	"fmt"

	c "github.com/okneniz/parsec/common"
)

// TODO : use for parsing another buffer?
// TODO : add unit tests too

type RunesBuffer struct {
	data     []rune
	position int

	// data string
	// positions []int (stack of last positions)
}

// newBuffer - make buffer which can read text on input
func NewRunesBuffer(str string) *RunesBuffer {
	b := new(RunesBuffer)
	b.data = []rune(str)
	b.position = 0
	return b
}

// Read - read next item, if greedy buffer keep position after reading.
func (b *RunesBuffer) Read(greedy bool) (rune, error) {
	if b.IsEOF() {
		return 0, c.EndOfFile
	}

	x := b.data[b.position]
	if greedy {
		b.position++
	}

	return x, nil
}

func (b *RunesBuffer) ReadAt(idx int) rune {
	return b.data[idx]
}

func (b *RunesBuffer) Size() int {
	return len(b.data)
}

func (b *RunesBuffer) String() string {
	return fmt.Sprintf("Buffer(%s, %d)", string(b.data), b.position)
}

// Seek - change buffer position
func (b *RunesBuffer) Seek(x int) {
	b.position = x
}

// Position - return current buffer position
func (b *RunesBuffer) Position() int {
	return b.position
}

// IsEOF - true if buffer ended
func (b *RunesBuffer) IsEOF() bool {
	return b.position >= len(b.data)
}
