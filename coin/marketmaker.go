package coin

import "fmt"

// MarketMaker 代表做市商，流动性挖矿提供者
type MarketMaker struct {
	Name  string
	CoinA *Coin
	CoinB *Coin
}

func (mm *MarketMaker) getCoin(name string) (*Coin, error) {
	switch name {
	case mm.CoinA.Name:
		return mm.CoinA, nil
	case mm.CoinB.Name:
		return mm.CoinB, nil
	default:
		return nil, fmt.Errorf("failed to get coin: illegal coin name")
	}
}
