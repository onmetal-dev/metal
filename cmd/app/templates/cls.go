package templates

import (
	"strings"
)

// cls generates a class string that combines all the classes contained in the argument list, potentially conditioning on
// the value of the previous argument.
// E.g. cls("input", inputErr, "input-error") will return "input input-error" if inputErr is not nil, otherwise "input"
func cls(args ...interface{}) string {
	classes := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == nil {
			i++ // Skip the next argument
			continue
		}
		switch v := args[i].(type) {
		case string:
			classes = append(classes, v)
		case bool:
			if i+1 < len(args) {
				if v {
					classes = append(classes, args[i+1].(string))
				}
				i++
			}
		case error:
			if i+1 < len(args) {
				if v != nil {
					classes = append(classes, args[i+1].(string))
				}
				i++
			}
		default:
			// non-nil something else, so add the next arg as long as it's a string
			if i+1 < len(args) {
				str, ok := args[i+1].(string)
				if ok {
					classes = append(classes, str)
				}
				i++
			}
		}
	}
	return strings.Join(classes, " ")
}
