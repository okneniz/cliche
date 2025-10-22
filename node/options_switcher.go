package node

import (
	"fmt"
	"sort"
	"strings"
)

type optionsSwitcher struct {
	enable  []ScanOption
	disable []ScanOption
	*base
}

func NewOptionsSwitcher(enable []ScanOption, disable []ScanOption) Node {
	const sep = ","

	sort.SliceStable(enable, func(i, j int) bool {
		return enable[i] < enable[j]
	})

	sort.SliceStable(disable, func(i, j int) bool {
		return disable[i] < disable[j]
	})

	enableStrings := make([]string, 0)
	for _, x := range enable {
		enableStrings = append(enableStrings, fmt.Sprintf("%v", x))
	}

	disableStrings := make([]string, 0)
	for _, x := range disable {
		disableStrings = append(disableStrings, fmt.Sprintf("%v", x))
	}

	enableKey := strings.Join(enableStrings, sep)
	disableKey := strings.Join(disableStrings, sep)

	var key string

	switch {
	case len(enable) > 0 && len(disable) == 0:
		key = fmt.Sprintf("options(enable=%s)", enableKey)
	case len(enable) == 0 && len(disable) > 0:
		key = fmt.Sprintf("options(disable=%s)", disableKey)
	default:
		key = fmt.Sprintf("options(enable=%s,disable=%s)", enableKey, disableKey)
	}

	return &optionsSwitcher{
		enable:  enable,
		disable: disable,
		base:    newBase(key),
	}
}

func (n *optionsSwitcher) Visit(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	optsPos := scanner.OptionsPosition()

	for _, opt := range n.enable {
		scanner.OptionsEnable(opt)
	}

	for _, opt := range n.disable {
		scanner.OptionsDisable(opt)
	}

	pos := scanner.Position()

	onMatch(n, from, from, true)
	n.base.VisitNested(scanner, input, from, to, onMatch)

	scanner.RewindOptions(optsPos)
	scanner.Rewind(pos)
}

func (n *optionsSwitcher) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}

func (n *optionsSwitcher) Copy() Node {
	return NewOptionsSwitcher(n.enable, n.disable)
}
