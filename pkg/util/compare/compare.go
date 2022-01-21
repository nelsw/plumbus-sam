package compare

import (
	"math"
	"plumbus/pkg/util/nums"
	"strconv"
	"strings"
)

const zero = "0"

func Strings(x, y string) bool {
	return fields(x, y) <= 0
}

func fields(x, y string) int {

	if x == y {
		return 0
	}

	xChunks := strings.Fields(x)
	yChunks := strings.Fields(y)

	xLen := len(xChunks)
	yLen := len(yChunks)

	var xStr, yStr string
	for i := 0; i < int(math.Min(float64(xLen), float64(yLen))); i++ {

		xStr = xChunks[i]
		yStr = yChunks[i]

		if xStr == yStr {
			continue
		}

		if nums.IsNotNumber(xStr) && nums.IsNotNumber(yStr) {
			return strings.Compare(xStr, yStr)
		} else if nums.IsNumber(xStr) && nums.IsNotNumber(yStr) {
			return -1
		} else if nums.IsNotNumber(xStr) && nums.IsNumber(yStr) {
			return 1
		} else if dif := digits(xStr, yStr); dif != 0 {
			return dif
		}
	}

	return intFun(xLen, yLen)
}

func digits(x, y string) int {

	if strings.HasPrefix(x, zero) && strings.HasPrefix(y, zero) {
		x = strings.Replace(x, zero, "", 1)
		y = strings.Replace(y, zero, "", 1)
		return digits(x, y)
	} else if strings.HasPrefix(x, zero) && !strings.HasPrefix(y, zero) {
		return -1
	} else if !strings.HasPrefix(x, zero) && strings.HasPrefix(y, zero) {
		return 1
	}

	return intFun(len(x), len(y), func() int {
		return intFun(rat(x), rat(y))
	})
}

func rat(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func intFun(x, y int, f ...func() int) int {
	if x < y {
		return -1
	} else if x > y {
		return 1
	} else if f != nil && len(f) > 0 {
		return f[0]()
	} else {
		return 0
	}
}
