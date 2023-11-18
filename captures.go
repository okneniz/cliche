package regular

import (
	"encoding/json"
	"fmt"
)

type captures struct {
	from  map[string]int
	to    map[string]int
	order []string
}

func newCaptures() *captures {
	return &captures{
		from:  make(map[string]int),
		to:    make(map[string]int),
		order: make([]string, 0),
	}
}

func (c *captures) String() string {
	x := make(map[string]interface{})
	x["from"] = c.from
	x["to"] = c.to
	x["order"] = c.order

	js, err := json.Marshal(x)
	if err != nil {
		return err.Error()
	}

	return string(js)
}

func (c *captures) IsEmpty() bool {
	return len(c.order) == 0
}

func (c *captures) Size() int {
	return len(c.order)
}

func (c *captures) From(name string, index int) {
	if _, exists := c.from[name]; exists {
		return
	}

	c.from[name] = index
	c.order = append(c.order, name)
}

func (c *captures) To(name string, index int) {
	if _, exists := c.from[name]; exists {
		c.to[name] = index
	}
}

func (c *captures) Delete(name string) {
	delete(c.from, name)
	delete(c.to, name)
	// TODO : maybe use map + slice for faster remove?
	c.order = remove[string](c.order, name)
}

func (c *captures) ToSlice() []Bounds {
	result := make([]Bounds, 0, len(c.to))

	var (
		start  int
		finish int
		exists bool
	)

	for _, name := range c.order {
		if start, exists = c.from[name]; !exists {
			break
		}

		if finish, exists = c.to[name]; !exists {
			break
		}

		result = append(result, Bounds{
			from: start,
			to:   finish,
		})
	}

	return result
}

func (c *captures) ToMap() map[string]Bounds {
	result := make(map[string]Bounds, len(c.to))

	var (
		start  int
		finish int
		exists bool
	)

	for _, name := range c.order {
		if start, exists = c.from[name]; !exists {
			break
		}

		if finish, exists = c.to[name]; !exists {
			break
		}

		result[name] = Bounds{
			from: start,
			to:   finish,
		}
	}

	return result
}

func remove[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}

	return l
}

type Bounds struct {
	from, to int
}

func (b Bounds) String() string {
	return fmt.Sprintf("%d-%d", b.from, b.to)
}
