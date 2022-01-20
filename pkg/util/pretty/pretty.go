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
	usd = accounting.Accounting{Symbol: "$", Precision: 2}
	dec = accounting.Accounting{Symbol: "", Precision: 2}
	num = accounting.Accounting{}
	pct = accounting.Accounting{Symbol: "%", Precision: 2, Format: "%v%s"}
)

func USD(v decimal.Decimal) string {
	return usd.FormatMoneyDecimal(v)
}

func Percent(v decimal.Decimal) string {
	return pct.FormatMoneyDecimal(v)
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
