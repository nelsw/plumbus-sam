package pretty

import (
	"encoding/json"
	"fmt"
	"github.com/leekchan/accounting"
	"github.com/shopspring/decimal"
)

var (
	num  = accounting.Accounting{}
	dec  = accounting.Accounting{Symbol: "", Precision: 2}
	usd0 = accounting.Accounting{Symbol: "$", Precision: 0}
	usd2 = accounting.Accounting{Symbol: "$", Precision: 2}
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

func Print(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "    ")
	fmt.Println(string(b))
}
