package pretty

import (
	"encoding/json"
	"fmt"
	"github.com/leekchan/accounting"
	"github.com/shopspring/decimal"
)

const (
	datetimeLayout = "2006-01-02 15:04:05"
	dateLayout     = "2006-01-02"
	timeLayout     = "15:04:05"
	zeroUSD        = "$0.00"
	zeroPCT        = "0.0%"
	zeroDEC        = "0.0"
)

var (
	usd0 = accounting.Accounting{Symbol: "$", Precision: 0}
	usd2 = accounting.Accounting{Symbol: "$", Precision: 2}
	dec  = accounting.Accounting{Symbol: "", Precision: 2}
	num  = accounting.Accounting{}
	pct0 = accounting.Accounting{Symbol: "%", Precision: 0, Format: "%v%s"}
	pct1 = accounting.Accounting{Symbol: "%", Precision: 1, Format: "%v%s"}
	pct2 = accounting.Accounting{Symbol: "%", Precision: 2, Format: "%v%s"}
)

func USD(v decimal.Decimal, b ...bool) string {
	if b != nil && len(b) > 0 && b[0] {
		return usd0.FormatMoneyDecimal(v)
	}
	return usd2.FormatMoneyDecimal(v)
}

func Percent(v decimal.Decimal, i int) string {
	if i == 0 {
		return pct0.FormatMoneyDecimal(v)
	} else if i == 1 {
		return pct1.FormatMoneyDecimal(v)
	} else {
		return pct2.FormatMoneyDecimal(v)
	}
}

func Int(v decimal.Decimal) string {
	return num.FormatMoneyDecimal(v)
}

func Decimal(v decimal.Decimal) string {
	return dec.FormatMoneyDecimal(v)
}

func Json(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

func PrintJson(v interface{}) {
	fmt.Println(Json(v))
}

func Print(v interface{}) {
	fmt.Println(Json(v))
}
