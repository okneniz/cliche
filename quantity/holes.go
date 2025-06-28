package quantity

type List interface {
	Size() int
	At(idx int) (Interface, bool)
}

func Get(x Interface, skips List) Interface {
	if x.Empty() {
		return x
	}

	if skips.Size() == 0 {
		return x
	}

	from := x.From()
	to := x.To()

	for i := skips.Size() - 1; i >= 0; i-- {
		if to < from {
			return Empty(x.From())
		}

		skip, _ := skips.At(i)

		// include
		if from < skip.From() && skip.To() < to {
			continue
		}

		// full cover
		if skip.From() <= from && to <= skip.To() {
			return Empty(x.From())
		}

		// right than but overlaped
		if from < skip.From() && to <= skip.To() {
			to = skip.From() - 1
			continue
		}

		// left than but overlaped
		if skip.From() <= from && skip.To() <= to {
			from = skip.To() + 1
			continue
		}
	}

	return Pair(from, to)
}
