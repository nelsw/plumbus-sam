package nums

import (
	"regexp"
)

var digitRegexp = regexp.MustCompile(`^[0-9]+$`)

func IsNumber(s string) bool {
	return digitRegexp.MatchString(s)
}
