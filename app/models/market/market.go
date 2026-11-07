package market

type EODData struct {
	Date          string  `json:"date"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	AdjustedClose float64 `json:"adjusted_close"`
	Volume        int64   `json:"volume"`
}

type LivePrice struct {
	Code          string  `json:"code"`
	Timestamp     int64   `json:"timestamp"`
	GMTOffset     int     `json:"gmtoffset"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        int64   `json:"volume"`
	PreviousClose float64 `json:"previousClose"`
	Change        float64 `json:"change"`
	ChangeP       float64 `json:"change_p"`
}

type SearchResult struct {
	Code     string `json:"Code"`
	Exchange string `json:"Exchange"`
	Name     string `json:"Name"`
	Type     string `json:"Type"`
	Country  string `json:"Country"`
	Currency string `json:"Currency"`
	ISIN     string `json:"ISIN,omitempty"`
}

type EODRequest struct {
	Symbol   string `query:"symbol" validate:"required"`
	Exchange string `query:"exchange"`
	From     string `query:"from"`
	To       string `query:"to"`
	Period   string `query:"period"`
	Order    string `query:"order"`
}

type LivePriceRequest struct {
	Symbol   string `query:"symbol" validate:"required"`
	Exchange string `query:"exchange"`
}

type SearchRequest struct {
	Query string `query:"q" validate:"required,min=1"`
	Limit int    `query:"limit"`
}
