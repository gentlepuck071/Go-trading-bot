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
	"github.com/saniales/golang-crypto-trading-bot/environment"
)

// TradeType represents a type of order, from trading fees point of view.
type TradeType string

const (
	// TakerTrade represents the "buy" order type.
	TakerTrade = "taker"
	// MakerTrade represents the "sell" order type.
	MakerTrade = "maker"
)

//ExchangeWrapper provides a generic wrapper for exchange services.
type ExchangeWrapper interface {
	Name() string                                                                                                // Gets the name of the exchange.
	GetTicker(market *environment.Market) (*environment.Ticker, error)                                           // Gets the updated ticker for a market.
	GetMarketSummary(market *environment.Market) (*environment.MarketSummary, error)                             // Gets the current market summary.
	GetOrderBook(market *environment.Market) (*environment.OrderBook, error)                                     // Gets the order(ASK + BID) book of a market.
	BuyLimit(market *environment.Market, amount float64, limit float64) (string, error)                          // Performs a limit buy action.
	SellLimit(market *environment.Market, amount float64, limit float64) (string, error)                         // Performs a limit sell action.
	CalculateTradingFees(market *environment.Market, amount float64, limit float64, orderType TradeType) float64 // Calculates the trading fees for an order on a specified market.
	CalculateWithdrawFees(market *environment.Market, amount float64) float64                                    // Calculates the withdrawal fees on a specified market.

<<<<<<< HEAD
	FeedConnect()                                                                                     // Connects to the feed of the exchange.
	SubscribeMarketSummaryFeed(market *environment.Market, onUpdate func(*environment.MarketSummary)) // Subscribes to the Market Summary Feed service.
<<<<<<< HEAD
	UnsubscribeMarketSummaryFeed(market *environment.Market)                                          // Unsubscribes from the Market Summary Feed service.                                                                                       // Disconnects from the feed
=======
	UnsubscribeMarketSummaryFeed(market *environment.Market)                                          // Unsubscribes from the Market Summary Feed service.
>>>>>>> bitfinex ws draft
=======
	FeedConnect()                                                                                    // Connects to the feed of the exchange.
	SubscribeMarketSummaryFeed(market *environment.Market, onUpdate func(environment.MarketSummary)) // Subscribes to the Market Summary Feed service.
	UnsubscribeMarketSummaryFeed(market *environment.Market)                                         // Unsubscribes from the Market Summary Feed service.
>>>>>>> adding bitfinex and binance support, adjusting other exchanges to match interface
}

// MarketNameFor gets the market name as seen by the exchange.
func MarketNameFor(m *environment.Market, wrapper ExchangeWrapper) string {
	return m.ExchangeNames[wrapper.Name()]
}
