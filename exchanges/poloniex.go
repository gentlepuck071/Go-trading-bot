// Copyright © 2017 Alessandro Sanino <saninoale@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package exchanges

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/pharrisee/poloniex-api"
	"github.com/saniales/golang-crypto-trading-bot/environment"
)

// PoloniexWrapper provides a Generic wrapper of the Poloniex API.
type PoloniexWrapper struct {
	api          *poloniex.Poloniex // access to Poloniex API
	tickerFeedUp bool               // if true, i am subscribing to market ticker.
}

// NewPoloniexWrapper creates a generic wrapper of the poloniex API.
func NewPoloniexWrapper(publicKey string, secretKey string) ExchangeWrapper {
	return PoloniexWrapper{
		api:          poloniex.NewWithCredentials(publicKey, secretKey),
		tickerFeedUp: false,
	}
}

// Name returns the name of the wrapped exchange.
func (wrapper PoloniexWrapper) Name() string {
	return "poloniex"
}

func (wrapper PoloniexWrapper) String() string {
	return wrapper.Name()
}

// GetMarkets gets all the markets info.
func (wrapper PoloniexWrapper) GetMarkets() ([]*environment.Market, error) {
	poloniexMarkets, err := wrapper.api.Currencies()
	if err != nil {
		return nil, err
	}
	wrappedMarkets := make([]*environment.Market, 0, len(poloniexMarkets))
	for _, market := range poloniexMarkets {
		if market.Disabled == 1 {
			name := strings.SplitN(market.Name, "/", 2)
			wrappedMarkets = append(wrappedMarkets, &environment.Market{
				Name:           market.Name,
				BaseCurrency:   name[1],
				MarketCurrency: name[0],
			})
		}
	}
	return wrappedMarkets, nil
}

// GetOrderBook gets the order(ASK + BID) book of a market.
func (wrapper PoloniexWrapper) GetOrderBook(market *environment.Market) (*environment.OrderBook, error) {
	poloniexOrderBook, err := wrapper.api.OrderBook(MarketNameFor(market, wrapper))
	if err != nil {
		return nil, err
	}

	var orderBook environment.OrderBook
	for _, order := range poloniexOrderBook.Bids {
		orderBook.Bids = append(orderBook.Bids, environment.Order{
			Quantity: decimal.NewFromFloat(order.Amount),
			Value:    decimal.NewFromFloat(order.Rate),
		})
	}
	for _, order := range poloniexOrderBook.Asks {
		orderBook.Asks = append(orderBook.Asks, environment.Order{
			Quantity: decimal.NewFromFloat(order.Amount),
			Value:    decimal.NewFromFloat(order.Rate),
		})
	}

	return &orderBook, nil
}

// BuyLimit performs a limit buy action.
func (wrapper PoloniexWrapper) BuyLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	orderNumber, err := wrapper.api.Buy(MarketNameFor(market, wrapper), amount, limit)
	return fmt.Sprint(orderNumber.OrderNumber), err
}

// SellLimit performs a limit sell action.
func (wrapper PoloniexWrapper) SellLimit(market *environment.Market, amount float64, limit float64) (string, error) {
	orderNumber, err := wrapper.api.Sell(MarketNameFor(market, wrapper), amount, limit)
	return fmt.Sprint(orderNumber.OrderNumber), err
}

// GetTicker gets the updated ticker for a market.
func (wrapper PoloniexWrapper) GetTicker(market *environment.Market) (*environment.Ticker, error) {
	poloniexTicker, err := wrapper.api.Ticker()
	if err != nil {
		return nil, err
	}
	ticker, exists := poloniexTicker[MarketNameFor(market, wrapper)]
	if !exists {
		return nil, errors.New("Market not found")
	}

	return &environment.Ticker{
		Last: decimal.NewFromFloat(ticker.Last),
		Bid:  decimal.NewFromFloat(ticker.Bid),
		Ask:  decimal.NewFromFloat(ticker.Ask),
	}, nil
}

// GetMarketSummary gets the current market summary.
func (wrapper PoloniexWrapper) GetMarketSummary(market *environment.Market) (*environment.MarketSummary, error) {
	poloniexSummaries, err := wrapper.api.Ticker()
	if err != nil {
		return nil, err
	}

	poloniexSummary, notExists := poloniexSummaries[MarketNameFor(market, wrapper)]
	if notExists {
		return nil, errors.New("Market not found")
	}

	return &environment.MarketSummary{
		Ask:    decimal.NewFromFloat(poloniexSummary.Ask),
		Bid:    decimal.NewFromFloat(poloniexSummary.Bid),
		Last:   decimal.NewFromFloat(poloniexSummary.Last),
		Volume: decimal.NewFromFloat(poloniexSummary.BaseVolume),
	}, nil
}

// CalculateTradingFees calculates the trading fees for an order on a specified market.
//
//     NOTE: In Binance fees are currently hardcoded.
func (wrapper PoloniexWrapper) CalculateTradingFees(market *environment.Market, amount float64, limit float64, orderType TradeType) float64 {
	// NOTE: possibility to use wrapper FeesInfo function.
	var feePercentage float64
	if orderType == MakerTrade {
		feePercentage = 0.0010
	} else if orderType == TakerTrade {
		feePercentage = 0.0020
	} else {
		panic("Unknown trade type")
	}

	return amount * limit * feePercentage
}

// CalculateWithdrawFees calculates the withdrawal fees on a specified market.
func (wrapper PoloniexWrapper) CalculateWithdrawFees(market *environment.Market, amount float64) float64 {
	panic("Not Implemented")
}

// FeedConnect connects to the feed of the poloniex websocket.
func (wrapper PoloniexWrapper) FeedConnect() {
	wrapper.api.StartWS()
}

var bindedTickers map[string]bool

// SubscribeMarketSummaryFeed subscribes to the Market Summary Feed service.
func (wrapper PoloniexWrapper) SubscribeMarketSummaryFeed(market *environment.Market, onUpdate func(environment.MarketSummary)) {
	subTicker := fmt.Sprintf("ticker:%s", MarketNameFor(market, wrapper))
	if len(bindedTickers) == 0 {
		wrapper.api.Subscribe("ticker")

		wrapper.api.On("ticker", func(t poloniex.WSTicker) {
			for bindedTicker := range bindedTickers {

				if bindedTicker == t.Pair {
					wrapper.api.Emit(subTicker, t)
				}
			}
		})
		wrapper.tickerFeedUp = true
	}

	if _, exists := bindedTickers[MarketNameFor(market, wrapper)]; !exists {
		bindedTickers[MarketNameFor(market, wrapper)] = true

		wrapper.api.On(subTicker, func(t poloniex.WSTicker) {
			onUpdate(environment.MarketSummary{
				High:   decimal.NewFromFloat(t.DailyHigh),
				Low:    decimal.NewFromFloat(t.DailyLow),
				Last:   decimal.NewFromFloat(t.Last),
				Ask:    decimal.NewFromFloat(t.Ask),
				Bid:    decimal.NewFromFloat(t.Bid),
				Volume: decimal.NewFromFloat(t.BaseVolume),
			})
		})
	}

}

// UnsubscribeMarketSummaryFeed unsubscribes from the Market Summary Feed service.
func (wrapper PoloniexWrapper) UnsubscribeMarketSummaryFeed(market *environment.Market) {
	subTicker := fmt.Sprintf("ticker:%s", MarketNameFor(market, wrapper))
	wrapper.api.Off(subTicker, func() {})
	delete(bindedTickers, MarketNameFor(market, wrapper))

	if len(bindedTickers) == 0 {
		wrapper.api.Unsubscribe("ticker")
	}
}
