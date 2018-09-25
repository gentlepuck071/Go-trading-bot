package exchanges

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/shopspring/decimal"

	bitfinex "github.com/bitfinexcom/bitfinex-api-go/v1"
	"github.com/saniales/golang-crypto-trading-bot/environment"
)

// BitfinexWrapper provides a Generic wrapper of the Bitfinex API.
type BitfinexWrapper struct {
	api                 *bitfinex.Client
	websocketOn         bool
	unsubscribeChannels map[string]chan bool
	summaries           SummaryCache
}

// NewBitfinexWrapper creates a generic wrapper of the bittrex API.
func NewBitfinexWrapper(publicKey string, secretKey string) ExchangeWrapper {
	return BitfinexWrapper{
		api:                 bitfinex.NewClient().Auth(publicKey, secretKey),
		unsubscribeChannels: make(map[string]chan bool),
		summaries:           NewSummaryCache(),
		websocketOn:         false,
	}
}

// Name returns the name of the wrapped exchange.
func (wrapper BitfinexWrapper) Name() string {
	return "bitfinex"
}

func (wrapper BitfinexWrapper) String() string {
	return wrapper.Name()
}

// GetMarkets gets all the markets info.
func (wrapper BitfinexWrapper) GetMarkets() ([]*environment.Market, error) {
	bitfinexMarkets, err := wrapper.api.Pairs.All()
	if err != nil {
		return nil, err
	}

	wrappedMarkets := make([]*environment.Market, len(bitfinexMarkets))
	for i, pair := range bitfinexMarkets {
		quote, base := pair[0:2], pair[3:5]
		wrappedMarkets[i] = &environment.Market{
			Name:           pair,
			BaseCurrency:   base,
			MarketCurrency: quote,
		}
	}

	return wrappedMarkets, nil
}

// GetOrderBook gets the order(ASK + BID) book of a market.
func (wrapper BitfinexWrapper) GetOrderBook(market *environment.Market) (*environment.OrderBook, error) {
	bitfinexOrderBook, err := wrapper.api.OrderBook.Get(MarketNameFor(market, wrapper), 0, 0, false)
	if err != nil {
		return nil, err
	}

	var orderBook environment.OrderBook
	for _, order := range bitfinexOrderBook.Bids {
		amount, _ := decimal.NewFromString(order.Amount)
		rate, _ := decimal.NewFromString(order.Rate)

		ts, err := order.ParseTime()
		if err != nil {
			ts = new(time.Time)
		}

		orderBook.Asks = append(orderBook.Asks, environment.Order{
			Quantity:  amount,
			Value:     rate,
			Timestamp: *ts,
		})
	}
	for _, order := range bitfinexOrderBook.Asks {
		amount, _ := decimal.NewFromString(order.Amount)
		rate, _ := decimal.NewFromString(order.Rate)

		ts, err := order.ParseTime()
		if err != nil {
			ts = new(time.Time)
		}

		orderBook.Bids = append(orderBook.Bids, environment.Order{
			Quantity:  amount,
			Value:     rate,
			Timestamp: *ts,
		})
	}

	return &orderBook, nil
}

// BuyLimit performs a limit buy action.
//
// NOTE: In bitfinex buy and sell orders behave the same (the go bitfinex api automatically puts it on correct side)
func (wrapper BitfinexWrapper) BuyLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	amount = math.Abs(amount)
	return wrapper.createOrder(market, bitfinex.OrderTypeLimit, amount, limit)
}

// SellLimit performs a limit sell action.
//
// NOTE: In bitfinex buy and sell orders behave the same (the go bitfinex api automatically puts it on correct side)
func (wrapper BitfinexWrapper) SellLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	amount = -math.Abs(amount) // a sell is a buy with negative amount.
	return wrapper.createOrder(market, bitfinex.OrderTypeLimit, amount, limit)
}

// BuyMarket performs a limit buy action.
//
// NOTE: In bitfinex buy and sell orders behave the same (the go bitfinex api automatically puts it on correct side)
func (wrapper BitfinexWrapper) BuyMarket(market *environment.Market, amount float64) (string, error) {
	amount = math.Abs(amount)
	return wrapper.createOrder(market, bitfinex.OrderTypeMarket, amount, 0)
}

// SellMarket performs a limit sell action.
//
// NOTE: In bitfinex buy and sell orders behave the same (the go bitfinex api automatically puts it on correct side)
func (wrapper BitfinexWrapper) SellMarket(market *environment.Market, amount float64) (string, error) {
	amount = -math.Abs(amount) // a sell is a buy with negative amount.
	return wrapper.createOrder(market, bitfinex.OrderTypeMarket, amount, 0)
}

// createOrder creates a generic order.
//
// NOTE: In bitfinex buy and sell orders behave the same (in sell the amount is negative)
func (wrapper BitfinexWrapper) createOrder(market *environment.Market, orderType string, amount float64, price float64) (string, error) {
	orderNumber, err := wrapper.api.Orders.Create(MarketNameFor(market, wrapper), amount, price, bitfinex.OrderTypeExchangeLimit)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(orderNumber.ID), nil
}

// GetTicker gets the updated ticker for a market.
func (wrapper BitfinexWrapper) GetTicker(market *environment.Market) (*environment.Ticker, error) {
	bitfinexTicker, err := wrapper.api.Ticker.Get(MarketNameFor(market, wrapper))
	if err != nil {
		return nil, err
	}

	last, _ := decimal.NewFromString(bitfinexTicker.LastPrice)
	ask, _ := decimal.NewFromString(bitfinexTicker.Ask)
	bid, _ := decimal.NewFromString(bitfinexTicker.Bid)

	return &environment.Ticker{
		Last: last,
		Bid:  bid,
		Ask:  ask,
	}, nil
}

// GetMarketSummary gets the current market summary.
func (wrapper BitfinexWrapper) GetMarketSummary(market *environment.Market) (*environment.MarketSummary, error) {
	ret, exists := wrapper.summaries.Get(market)
	if !exists || !wrapper.websocketOn {
		bitfinexSummary, err := wrapper.api.Ticker.Get(MarketNameFor(market, wrapper))
		if err != nil {
			return nil, err
		}

		high, _ := decimal.NewFromString(bitfinexSummary.High)
		low, _ := decimal.NewFromString(bitfinexSummary.Low)
		volume, _ := decimal.NewFromString(bitfinexSummary.Volume)
		bid, _ := decimal.NewFromString(bitfinexSummary.Bid)
		ask, _ := decimal.NewFromString(bitfinexSummary.Ask)

		ret = &environment.MarketSummary{
			High:   high,
			Low:    low,
			Volume: volume,
			Bid:    bid,
			Ask:    ask,
			Last:   ask, // TODO: find a better way for last value, if any
		}

		wrapper.summaries.Set(market, ret)
		return ret, nil
	}

	if !summaryLoaded {
		return nil, errors.New("Summary not loaded")
	}

	return ret, nil
}

// GetCandles gets the candle data from the exchange.
func (wrapper BitfinexWrapper) GetCandles(market *environment.Market) ([]environment.CandleStick, error) {
	panic("Not supported in V1")
}

// GetBalance gets the balance of the user of the specified currency.
func (wrapper BitfinexWrapper) GetBalance(symbol string) (*decimal.Decimal, error) {
	bitfinexBalances, err := wrapper.api.Balances.All()
	if err != nil {
		return nil, err
	}

	for _, bitfinexBalance := range bitfinexBalances {
		if bitfinexBalance.Currency == symbol {
			ret, err := decimal.NewFromString(bitfinexBalance.Available)
			if err != nil {
				return nil, err
			}

			return &ret, nil
		}
	}

	return nil, errors.New("Symbol not found")
}

// CalculateTradingFees calculates the trading fees for an order on a specified market.
//
//     NOTE: In Bitfinex fees are currently hardcoded.
func (wrapper BitfinexWrapper) CalculateTradingFees(market *environment.Market, amount float64, limit float64, orderType TradeType) float64 {
	var feePercentage float64
	if orderType == MakerTrade {
		feePercentage = 0.0010 // 0.1%
	} else if orderType == TakerTrade {
		feePercentage = 0.0020 // 0.2%
	} else {
		panic("Unknown trade type")
	}

	return amount * limit * feePercentage
}

// CalculateWithdrawFees calculates the withdrawal fees on a specified market.
func (wrapper BitfinexWrapper) CalculateWithdrawFees(market *environment.Market, amount float64) float64 {
	panic("Not Implemented")
}

// FeedConnect connects to the feed of the exchange.
func (wrapper BitfinexWrapper) FeedConnect() {
	err := wrapper.api.WebSocket.Connect()
	if err != nil {
		panic(err)
	}

	wrapper.websocketOn = true
}

// SubscribeMarketSummaryFeed subscribes to the Market Summary Feed service.
//
//     NOTE: Content of result array
//     BID	float	Price of last highest bid
//     BID_SIZE	float	Size of the last highest bid
//     ASK	float	Price of last lowest ask
//     ASK_SIZE	float	Size of the last lowest ask
//     DAILY_CHANGE	float	Amount that the last price has changed since yesterday
//     DAILY_CHANGE_PERC	float	Amount that the price has changed expressed in percentage terms
//     LAST_PRICE	float	Price of the last trade.
//     VOLUME	float	Daily volume
//     HIGH	float	Daily high
//     LOW	float	Daily low
func (wrapper BitfinexWrapper) SubscribeMarketSummaryFeed(market *environment.Market) {
	results := make(chan []float64)
	tickerKey := MarketNameFor(market, wrapper)

	wrapper.api.WebSocket.AddSubscribe("ticker", tickerKey, results)

	wrapper.unsubscribeChannels[MarketNameFor(market, wrapper)] = make(chan bool, 1)

	go func(results <-chan []float64, tickerKey string) {
		for {
			select {
			case values := <-results:
				wrapper.summaries.Set(market, &environment.MarketSummary{
					Bid:    decimal.NewFromFloat(values[0]),
					Ask:    decimal.NewFromFloat(values[2]),
					Volume: decimal.NewFromFloat(values[7]),
					High:   decimal.NewFromFloat(values[8]),
					Low:    decimal.NewFromFloat(values[9]),
				})
			case <-wrapper.unsubscribeChannels[tickerKey]:
				close(wrapper.unsubscribeChannels[tickerKey])
				delete(wrapper.unsubscribeChannels, tickerKey)
				if len(wrapper.unsubscribeChannels) == 0 {
					wrapper.api.WebSocket.ClearSubscriptions()
				}
				return
			}
		}
	}(results, tickerKey)
}

// UnsubscribeMarketSummaryFeed unsubscribes from the Market Summary Feed service.
func (wrapper BitfinexWrapper) UnsubscribeMarketSummaryFeed(market *environment.Market) {
	if _, subscribedToFeed := wrapper.unsubscribeChannels[MarketNameFor(market, wrapper)]; subscribedToFeed {
		wrapper.unsubscribeChannels[MarketNameFor(market, wrapper)] <- true
	}
}
