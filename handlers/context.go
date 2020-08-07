package handlers

import "github.com/alpacahq/alpaca-trade-api-go/alpaca"

type TradingContext struct {
	Client                 *alpaca.Client
	AssetKey               string
	Quote                  Quote
	Position               Position
	MaxShareToHold         int32
	ThresholdToFollow      int32
	DefaultQuantityToTrade int32
}

func NewTradingContext(client *alpaca.Client, symbol string,
	maxShareToHold int32, thresholdToFollow int32, defaultTradeQuantity int32) *TradingContext {
	quote := NewQuote()
	position := NewPosition()

	return &TradingContext{
		Client:                 client,
		Quote:                  quote,
		Position:               position,
		AssetKey:               symbol,
		MaxShareToHold:         maxShareToHold,
		ThresholdToFollow:      thresholdToFollow,
		DefaultQuantityToTrade: defaultTradeQuantity,
	}
}
