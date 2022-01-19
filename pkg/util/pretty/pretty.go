package pretty

import (
	"encoding/json"
	"fmt"
	"github.com/leekchan/accounting"
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

func USD(v interface{}) string {
	return usd.FormatMoney(v)
}

func Percent(v interface{}) string {
	return pct.FormatMoney(v)
}

func Int(v interface{}) string {
	return num.FormatMoney(v)
}

func Decimal(v interface{}) string {
	return dec.FormatMoney(v)
}

func Json(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

func PrintJson(v interface{}) {
	fmt.Println(Json(v))
}
