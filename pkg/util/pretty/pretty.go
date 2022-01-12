package pretty

import "github.com/leekchan/accounting"

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

func FormatUSD(v interface{}) string {
	return usd.FormatMoney(v)
}

func FormatPercentage(v interface{}) string {
	return pct.FormatMoney(v)
}

func FormatInt(v interface{}) string {
	return num.FormatMoney(v)
}

func FormatDecimal(v interface{}) string {
	return dec.FormatMoney(v)
}
