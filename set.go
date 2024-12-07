package cliche

type Set map[string]struct{}

func newSet(items ...string) Set {
	d := make(Set)
	for _, x := range items {
		d.add(x)
	}
	return d
}

func (d Set) add(str string) {
	d[str] = struct{}{}
}

func (d Set) merge(other Set) Set {
	for key, value := range other {
		d[key] = value
	}

	return d
}

func (d Set) Slice() []string {
	result := make([]string, len(d))
	i := 0
	for key := range d {
		result[i] = key
		i++
	}
	return result
}

