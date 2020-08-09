package hftish

import (
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"hftish-go/logging"
	"log"
	"math"
)

var logger *log.Logger

func init() {
	logger = logging.GetLogger()
}

type Quote struct {
	PrevBid     float32
	PrevAsk     float32
	PrevSpread  float32
	Bid         float32
	Ask         float32
	Spread      float32
	BidSize     int32
	AskSize     int32
	Traded      bool
	LevelChange int32
	Time        int64
}

func NewQuote() Quote {
	return Quote{
		LevelChange: 1,
		Traded:      true,
		//all others take default value.
	}
}

func (quote *Quote) Reset() {
	// called when a level change happens
	quote.Traded = false
	quote.LevelChange += 1
}

func (quote *Quote) Update(qData alpaca.StreamQuote) {
	if quote.Time == 0 {
		quote.Time = qData.Timestamp
		quote.Bid = qData.BidPrice
		quote.BidSize = qData.BidSize
		quote.Ask = qData.AskPrice
		quote.AskSize = qData.AskSize
		logger.Printf("Init Quote -- Bid: %f, Ask: %f", quote.Bid, quote.Ask)
		return
	}

	// update bid and ask sizes and timestamp.
	quote.BidSize = qData.BidSize
	quote.AskSize = qData.AskSize

	b1 := quote.Bid != qData.BidPrice
	b2 := quote.Ask != qData.AskPrice
	spread := round(qData.AskPrice-qData.BidPrice, 2)

	logger.Printf("checking -- Bid: %t, Ask: %t, spread: %f", b1, b2, spread)

	// check if there has been a level change
	if b1 && b2 && spread == 0.01 {
		quote.PrevBid = quote.Bid
		quote.PrevAsk = quote.Ask
		quote.Bid = qData.BidPrice
		quote.Ask = qData.AskPrice
		quote.Time = qData.Timestamp
		//update spread.
		s1 := round(quote.PrevAsk-quote.PrevBid, 3)
		s2 := round(quote.Ask-quote.Bid, 3)

		quote.PrevSpread = s1
		quote.Spread = s2

		logger.Printf("spread -- prev: %f, now: %f", s1, s2)

		// if change is from one penny spread level to a different penny
		// spread level, then initialize for new level
		if quote.PrevSpread == 0.01 {
			quote.Reset()
		}
	}
}

func NewPosition() Position {
	return Position{
		OrderFilledAmount: make(map[string]int32),
	}
}

type Position struct {
	OrderFilledAmount map[string]int32
	PendingBuyShares  int32
	PendingSellShares int32
	TotalShares       int32
}

func (p *Position) ToString() string {
	return fmt.Sprintf("Position{ TotalShare: %d, Pending Buy: %s, Pending Sell: %s, Orderfilled: %s}",
		p.TotalShares, p.PendingBuyShares, p.PendingSellShares, p.OrderFilledAmount)
}

func (p *Position) UpdatePendingBuyShares(quantity int32) {
	p.PendingBuyShares += quantity
}

func (p *Position) UpdatePendingSellShares(quantity int32) {
	p.PendingSellShares += quantity
}

func (p *Position) UpdateTotalShares(quantity int32) {
	p.TotalShares += quantity
}

func (p *Position) UpdateFilledAmount(orderId string, newAmount int32, side string) {
	oldAmount := p.OrderFilledAmount[orderId] // oldAmount default to 0 so there is no need to check existence
	if newAmount > oldAmount {
		if side == "buy" {
			p.UpdatePendingBuyShares(oldAmount - newAmount) //filled, means not pending anymore
			p.UpdateTotalShares(newAmount - oldAmount)
		} else { // sell
			p.UpdatePendingSellShares(oldAmount - newAmount)
			p.UpdateTotalShares(oldAmount - newAmount)
		}
		p.OrderFilledAmount[orderId] = newAmount
	}
}

func (p *Position) RemovePendingOrder(orderId string, side string) {
	oldAmount, ok := p.OrderFilledAmount[orderId]
	if ok {
		if side == "buy" {
			p.UpdatePendingBuyShares(oldAmount - 100) // why 100?
		} else {
			p.UpdatePendingSellShares(oldAmount - 100)
		}
		delete(p.OrderFilledAmount, orderId)
	}
}

//round the float to the decimal.
func round(num1 float32, round int) float32 {
	scale := 10
	for i := 0; i < round; i++ {
		scale = scale * 10
	}
	return float32(math.Round(float64(num1)*float64(scale)) / float64(scale))
}
