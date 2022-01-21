package nums

import (
	"strconv"
)

func IsNumber(s string) bool {
	return isNumber(s)
}

func IsNotNumber(s string) bool {
	return !IsNumber(s)
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
