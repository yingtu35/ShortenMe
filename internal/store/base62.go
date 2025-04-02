package store

import (
	"math"
)

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

// Decode converts a base62 string to number
func DecodeBase62(str string) int64 {
	var num int64
	for i, char := range []byte(str) {
		for j, b := range []byte(base62Chars) {
			if b == char {
				num += int64(j) * int64(math.Pow(62, float64(len(str)-i-1)))
			}
		}
	}
	return num
}
