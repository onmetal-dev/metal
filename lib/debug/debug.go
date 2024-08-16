package debug

import "encoding/json"

func PrettyJSON(i interface{}) string {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
