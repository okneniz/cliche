package node

import "fmt"

type validator func(string, Node) error

func checkEmptyClasses(expression string, start Node) error {
	err := fmt.Errorf("empty char-class: /%s/", expression)
	found := false

	Traverse(start, func(n Node) bool {
		switch x := n.(type) {
		case *class:
			found = x.table.Empty()
		case *negativeClass:
			found = x.table.Empty()
		case Alternation:
			for _, variant := range x.GetVariants() {
				if err = checkEmptyClasses(expression, variant); err != nil {
					found = true
				}
			}
		case Container:
			if err = checkEmptyClasses(expression, x.GetValue()); err != nil {
				found = true
			}
		}

		return found
	})

	if found {
		return err
	}

	return nil
}

var defaultValidators = []validator{
	checkEmptyClasses,
}

func Validate(expression string, start Node) error {
	for _, validate := range defaultValidators {
		if err := validate(expression, start); err != nil {
			return err
		}
	}

	return nil
}
