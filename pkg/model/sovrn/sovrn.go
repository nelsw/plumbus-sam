package sovrn

type Entity struct {
	UTM         string  `json:"UTM"`
	Revenue     float64 `json:"Revenue"`
	Impressions int     `json:"Impressions"`
	Sessions    int     `json:"Sessions"`
	CTR         float64 `json:"CTR"`
	PageViews   int     `json:"PageViews"`
}
