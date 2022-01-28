package nums

import (
	"reflect"
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

func Float64(v interface{}) float64 {
	if v == nil {
		return 0
	} else if k := reflect.ValueOf(v).Kind(); k == reflect.String {
		if s := v.(string); s == "" {
			return 0
		} else if f, err := strconv.ParseFloat(s, 64); err != nil {
			return 0
		} else {
			return f
		}
	} else if k == reflect.Float64 {
		return v.(float64)
	} else {
		return 0
	}
}
