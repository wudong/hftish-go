package main

import (
	"flag"
	"fmt"
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/alpacahq/alpaca-trade-api-go/stream"
	hftish "hftish-go"
	"hftish-go/handlers"
	"log"
	"os"
)

const (
	ApiSecret           = "3ACKbLvfY/ZgTBCwVGycr62Hc1/jJyRWj4bi/aqa"
	ApiKey              = "PKJ6ANPZLVR4F3NUS4S5"
	DEFAULT_BASE_URL    = "https://paper-api.alpaca.markets"
	DEFAULT_ASSET       = "DB"
	MAX_SHARE_TO_HOLD   = 1000
	THRESHOLD_TO_FOLLOW = 10
	DEFAULT_QUANTITY    = 10
)

var logger = log.New(os.Stdout, "main", log.LstdFlags)

func init() {
	if env := os.Getenv(common.EnvApiKeyID); env == "" {
		os.Setenv(common.EnvApiKeyID, ApiKey)
	}

	if env := os.Getenv(common.EnvApiSecretKey); env == "" {
		os.Setenv(common.EnvApiSecretKey, ApiSecret)
	}
}

func main() {

	asset := flag.String("asset", DEFAULT_ASSET, "Specify which asset you want to trade with.")
	url := flag.String("url", DEFAULT_BASE_URL, "Specify the API URL.")
	maxShare := flag.Int("maxShare", MAX_SHARE_TO_HOLD, "The max number of share we want to hold.")
	threshold := flag.Int("threshold", THRESHOLD_TO_FOLLOW, "A big enough trade for us to follow.")
	tradeQuantity := flag.Int("defaultTradeQuantity", DEFAULT_QUANTITY, "The trade quantity for each trade to be executed.")
	flag.Parse()

	alpaca.SetBaseUrl(*url)

	qc := fmt.Sprintf("Q.%s", *asset)
	tc := fmt.Sprintf("T.%s", *asset)

	alpacaClient := alpaca.NewClient(common.Credentials())
	acct, err := alpacaClient.GetAccount()
	if err != nil {
		panic(err)
	}
	logger.Println("Account: ", *acct)

	context := hftish.NewTradingContext(alpacaClient, *asset, int32(*maxShare), int32(*threshold), int32(*tradeQuantity))
	logger.Printf("Context{asset: %s, qty: %d, max: %d, follow: %d}\n", context.AssetKey, context.DefaultQuantityToTrade, context.MaxShareToHold, context.ThresholdToFollow)

	if err := stream.Register(alpaca.TradeUpdates, func(msg interface{}) {
		tradeUpdateMsg := msg.(alpaca.TradeUpdate)
		handlers.TradeUpdateHandler(context, tradeUpdateMsg)
	}); err != nil {
		panic(err)
	}

	if err := stream.Register(qc, func(msg interface{}) {
		quoteMsg := msg.(alpaca.StreamQuote)
		handlers.QuoteHandler(context, quoteMsg)
	}); err != nil {
		panic(err)
	}

	if err := stream.Register(tc, func(msg interface{}) {
		tradeMsg := msg.(alpaca.StreamTrade)
		handlers.TradeHandler(context, tradeMsg)
	}); err != nil {
		panic(err)
	}

	select {}

}
