package global

import "encoding/json"

func DoubleEscapeCheck(body []byte) []byte {
	var decoded string
	err := json.Unmarshal(body, &decoded)
	if err != nil {
		return body
	}
	return []byte(decoded)
}
