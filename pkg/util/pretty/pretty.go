package pretty

import (
	"encoding/json"
	"fmt"
	"github.com/leekchan/accounting"
	"github.com/shopspring/decimal"
)

var (
	num  = accounting.Accounting{}
	usd0 = accounting.Accounting{Symbol: "$", Precision: 0}
	usd2 = accounting.Accounting{Symbol: "$", Precision: 2}
	pct0 = accounting.Accounting{Symbol: "%", Precision: 0, Format: "%v%s"}
	pct1 = accounting.Accounting{Symbol: "%", Precision: 1, Format: "%v%s"}
	pct2 = accounting.Accounting{Symbol: "%", Precision: 2, Format: "%v%s"}
)

func USD(v decimal.Decimal, b ...bool) string {
	if b != nil && len(b) > 0 && b[0] {
		return usd0.FormatMoney(v)
	}
	return usd2.FormatMoney(v)
}

func Percent(v decimal.Decimal, ii ...int) string {
	var i int
	if ii != nil && len(ii) > 0 {
		i = ii[0]
	}
	if i == 0 {
		return pct0.FormatMoney(v)
	} else if i == 1 {
		return pct1.FormatMoney(v)
	} else {
		return pct2.FormatMoney(v)
	}
}

func Int(v decimal.Decimal) string {
	return num.FormatMoney(v)
}

func Print(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "    ")
	fmt.Println(string(b))
}
