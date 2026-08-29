package httpx

import "encoding/json"

// itoa 整数转字符串。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// jsonUnmarshalInto 宽松 JSON 解析（字符串入参）。
func jsonUnmarshalInto(raw string, v any) error {
	return json.Unmarshal([]byte(raw), v)
}
