package store

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Encode converts a number to base62 string
func EncodeBase62(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var result []byte
	for num > 0 {
		result = append([]byte{base62Chars[num%62]}, result...)
		num /= 62
	}
	return string(result)
}
