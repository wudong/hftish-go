package handlers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/shopspring/decimal"
	hftish "hftish-go"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "handlers", log.LstdFlags)

func placeOrder(context *hftish.TradingContext, assetKey string, price decimal.Decimal, side alpaca.Side) (*alpaca.Order, error) {
	order := alpaca.PlaceOrderRequest{
		AssetKey:    &assetKey,
		Side:        side,
		Type:        alpaca.Limit,
		TimeInForce: alpaca.IOC,
		Qty:         decimal.NewFromInt(int64(context.DefaultQuantityToTrade)),
		LimitPrice:  &price,
	}

	placeOrder, err := context.Client.PlaceOrder(order)
	if err != nil {
		logger.Fatal("Submit order failed", err)
		return nil, err
	}
	return placeOrder, nil
}

// we got an update on one of the orders we submitted, we need to update
// our Position with the new information
func TradeUpdateHandler(context *hftish.TradingContext, msg alpaca.TradeUpdate) {
	logger.Printf("%s event received for order %s.\n", msg.Event, msg.Order.ID)
	event := msg.Event

	orderId := msg.Order.ID
	orderSide := string(msg.Order.Side)

	switch {
	case event == "fill":
		if msg.Order.Side == alpaca.Buy {
			context.Position.UpdateTotalShares(int32(msg.Order.FilledQty.IntPart()))
		} else {
			context.Position.UpdateTotalShares(-1 * int32(msg.Order.FilledQty.IntPart()))
		}
		context.Position.RemovePendingOrder(orderId, orderSide)
	case event == "partial_fill":
		context.Position.UpdateFilledAmount(orderId, int32(msg.Order.FilledQty.IntPart()), orderSide)
	case event == "canceled" || event == "rejected":
		context.Position.RemovePendingOrder(orderId, orderSide)
	default:
		logger.Print("Not recognized event received: ", msg.Event)
	}

	logger.Println("Current Position:", context.Position)

}

func QuoteHandler(context *hftish.TradingContext, msg alpaca.StreamQuote) {
	logger.Println("Quote Received:", msg.Symbol, msg.BidPrice, msg.BidSize, msg.AskPrice, msg.AskSize)
	context.Quote.Update(msg)
}

//
func TradeHandler(context *hftish.TradingContext, msg alpaca.StreamTrade) {
	logger.Println("Trade Received:", msg.Symbol, msg.Price, msg.Size)
	if context.Quote.Traded { // have we already traded on this.
		return
	}

	if msg.Timestamp <= context.Quote.Time+50 { //the trade came too close o
		return
	}

	if msg.Size > context.ThresholdToFollow {
		// the trade was large enough to follow. so we check to to see if we are ready to trade.
		// we also check to see that the bid vs ask quantities (order book imbalance)
		// indicate a moment in that direction.
		logger.Println("Following the trade:", msg.Symbol, msg.Price)

		assetKey := context.AssetKey
		price := decimal.NewFromFloat32(msg.Price)

		if msg.Price == context.Quote.Ask && // we are buying.
			float32(context.Quote.BidSize) > 1.8*float32(context.Quote.AskSize) &&
			context.Position.TotalShares+context.Position.PendingBuyShares < context.MaxShareToHold {
			//submit our buy at the ask price
			order, err := placeOrder(context, assetKey, price, alpaca.Buy)
			if err != nil {
				return
			}

			context.Position.UpdatePendingBuyShares(context.DefaultQuantityToTrade)
			context.Position.OrderFilledAmount[order.ID] = 0
			context.Quote.Traded = true
			logger.Printf("Buy at: %s\n", price)
		} else if msg.Price == context.Quote.Bid && // we are selling.
			float32(context.Quote.AskSize) > 1.8*float32(context.Quote.BidSize) &&
			context.Position.TotalShares-context.Position.PendingSellShares >= context.DefaultQuantityToTrade {

			order, err := placeOrder(context, assetKey, price, alpaca.Sell)
			if err != nil {
				return
			}

			context.Position.UpdatePendingSellShares(context.DefaultQuantityToTrade)
			context.Position.OrderFilledAmount[order.ID] = 0
			context.Quote.Traded = true
			logger.Printf("Sell at: %s\n", price)
		} else {
			logger.Printf("Signal not met, ignore the trade")
		}
	}
}
