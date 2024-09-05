package templates

// cls generates a class string that combines a base class with conditional classes based on the provided pairs.
// E.g. cls("input", inputErr, "input-error") will return "input input-error" if inputErr is not nil, otherwise "input"
func cls(class string, pairs ...interface{}) string {
	for i := 0; i < len(pairs); i += 2 {
		if i+1 >= len(pairs) {
			break // Ensure we have both value and valueClass
		}
		value := pairs[i]
		valueClass := pairs[i+1].(string)

		if value != nil {
			switch v := value.(type) {
			case bool:
				if v {
					class += " " + valueClass
				}
			default:
				class += " " + valueClass
			}
		}
	}
	return class
}
