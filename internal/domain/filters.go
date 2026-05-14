package domain

// MarketFilter описывает параметры фильтрации и пагинации рынков.
type MarketFilter struct {
	Status *string
	SortBy string
	Page   int
	Limit  int
}

// OrderFilter описывает параметры фильтрации и пагинации ордеров.
type OrderFilter struct {
	Status *string
	Page   int
	Limit  int
}

// TradeFilter описывает параметры фильтрации и пагинации сделок.
type TradeFilter struct {
	MarketID *string
	Page     int
	Limit    int
}
