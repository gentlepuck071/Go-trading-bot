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
	"sync"

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

	FeedConnect()                                            // Connects to the feed of the exchange.
	SubscribeMarketSummaryFeed(market *environment.Market)   // Subscribes to the Market Summary Feed service.
	UnsubscribeMarketSummaryFeed(market *environment.Market) // Unsubscribes from the Market Summary Feed service.
}

// SummaryCache represents a local summary cache for every exchange. To allow dinamic polling from multiple sources (REST + Websocket)
type SummaryCache struct {
	mutex    *sync.RWMutex
	internal map[*environment.Market]*environment.MarketSummary
}

// NewSummaryCache creates a new SummaryCache Object
func NewSummaryCache() SummaryCache {
	return SummaryCache{
		mutex:    &sync.RWMutex{},
		internal: make(map[*environment.Market]*environment.MarketSummary),
	}
}

// Set sets a value for the specified key.
func (sc *SummaryCache) Set(market *environment.Market, summary *environment.MarketSummary) *environment.MarketSummary {
	sc.mutex.Lock()
	old := sc.internal[market]
	sc.internal[market] = summary
	sc.mutex.Unlock()
	return old
}

// Get gets the value for the specified key.
func (sc *SummaryCache) Get(market *environment.Market) (*environment.MarketSummary, bool) {
	sc.mutex.RLock()
	ret, isSet := sc.internal[market]
	sc.mutex.RUnlock()
	return ret, isSet
}

// MarketNameFor gets the market name as seen by the exchange.
func MarketNameFor(m *environment.Market, wrapper ExchangeWrapper) string {
	return m.ExchangeNames[wrapper.Name()]
}
