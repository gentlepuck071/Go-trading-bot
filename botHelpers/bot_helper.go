package botHelpers

import (
	"github.com/AlessandroSanino1994/gobot/environment"
	"github.com/AlessandroSanino1994/gobot/exchangeWrappers"
)

//InitExchange initialize a new ExchangeWrapper binded to the specified exchange provided.
func InitExchange(exchangeConfig environment.ExchangeConfig) exchangeWrappers.ExchangeWrapper {
	switch exchangeConfig.ExchangeName {
	case "bittrex":
		return exchangeWrappers.NewBittrexWrapper(exchangeConfig.PublicKey, exchangeConfig.SecretKey)
	case "poloniex":
		return nil
	default:
		return nil
	}
}

//InitMarkets uses ExchangeWrapper to find info about markets and initialize them.
func InitMarkets(exchange exchangeWrappers.ExchangeWrapper) (map[string]environment.Market, error) {
	markets, err := exchange.GetMarkets()
	if err != nil {
		return nil, err
	}

	marketMap := make(map[string]environment.Market, len(markets))
	for _, market := range markets {
		marketMap[market.Name] = market
	}

	return marketMap, nil
}
