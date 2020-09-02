package coin

import (
	"fmt"
	"strings"
)

// Coin represents a coin
type Coin struct {
	Name   string
	Amount float64
}

// Add 用来增加coin的数量
func (c *Coin) Add(value float64) {
	c.Amount += value
}

// Del 用来减少coin的数量
func (c *Coin) Del(value float64) {
	if value <= c.Amount {
		c.Amount -= value
	}
}

// CoinPair represents a pair of coins
type CoinPair struct {
	CoinA   *Coin
	CoinB   *Coin
	totalK  float64
	feeRate float64
	fees    float64

	makers map[string]*MarketMaker
}

// NewCoinPair returns a CoinPair instance
// Note: If name is eht-mtv，coin A'name is eth, coin B'name is mtv
func NewCoinPair(name string, A float64, B float64, fee float64, makerName string) (*CoinPair, error) {
	cp := new(CoinPair)
	names := strings.Split(name, "-")
	if len(names) != 2 {
		return nil, fmt.Errorf("faild to create new coin pair: illegal coin name")
	}

	coinA := &Coin{Name: names[0], Amount: A}
	coinB := &Coin{Name: names[1], Amount: B}
	cp.CoinA = coinA
	cp.CoinB = coinB
	cp.totalK = A * B
	cp.feeRate = fee
	cp.makers = make(map[string]*MarketMaker, 0)
	cp.addMaker(makerName, A, B)
	return cp, nil
}

// Send is used to exchange coins
func (cp *CoinPair) Send(from, to string, value float64) (float64, error) {
	fromCoin, err := cp.getCoin(from)
	if err != nil {
		return 0, err
	}
	toCoin, err := cp.getCoin(to)
	if err != nil {
		return 0, err
	}

	if fromCoin.Amount-value < 0 {
		return 0, fmt.Errorf("no enough coin to send")
	}

	// 扣除手续费
	fee := cp.feeRate * value
	value = value - fee

	// 计算应该得到的想要买的币的数量
	newAmountOfTo := cp.totalK / (fromCoin.Amount + value)
	getToCoin := newAmountOfTo - toCoin.Amount

	// 调整两个币的数量，保持平衡
	fromCoin.Add(value)
	toCoin.Del(getToCoin)

	// 检查下调整后的k是否变化
	if err := cp.checkk(); err != nil {
		// 如果变化，那么回滚并返回错误
		fromCoin.Del(value)
		toCoin.Add(getToCoin)
		return 0, err
	}

	// 手续费池里面增加一笔钱
	cp.shareFee(from, fee)
	return getToCoin, nil
}

// AddLiquid 流动性抵押
func (cp *CoinPair) AddLiquid(coinA string, valueA float64, coinB string, valueB float64, makerName string) error {
	// 如果增加的比率和现有的利率不等，那么返回错误
	if valueA/valueB != cp.CoinA.Amount/cp.CoinB.Amount {
		return fmt.Errorf("invalid ratio of coin pair")
	}

	// 修改资金池数量
	cp.CoinA.Add(valueA)
	cp.CoinB.Add(valueB)
	// 更新k值
	cp.updatek()

	// 查找下流动挖矿者是否存在，不存在新建一个账户
	cp.addMaker(makerName, valueA, valueB)
	return nil
}

// GetAddNum 用来计算当存入一种币时，需要存入对应的另一种币的数量
func (cp *CoinPair) GetAddNum(coinname string, value float64) (float64, error) {
	// 首先判断存入的是哪种币
	coin, err := cp.getCoin(coinname)
	if err != nil {
		return 0, err
	}

	// 获取另一种币
	var calCoin *Coin
	if coin.Name == cp.CoinA.Name {
		calCoin = cp.CoinB
	} else {
		calCoin = cp.CoinA

	}

	// 计算两个币的汇率
	exRate := calCoin.Amount / coin.Amount
	// 返回另一种币需要存入的量
	return exRate * value, nil

}

// shareFee 用来评分手续费的
func (cp *CoinPair) shareFee(coinname string, fee float64) error {
	// 首先计算一共有多少投入
	var totalCoins float64
	for _, v := range cp.makers {
		makerCoin, err := v.getCoin(coinname)
		if err != nil {
			return err
		}
		totalCoins += makerCoin.Amount
	}

	// 再按照每个旷工的投入比例分配奖励
	for _, v := range cp.makers {
		makerCoin, err := v.getCoin(coinname)
		if err != nil {
			return err
		}
		getFee := makerCoin.Amount / totalCoins * fee
		makerCoin.Amount += getFee
	}

	return nil
}

// GetMaker 根据maker的名字获取maker对象
func (cp *CoinPair) GetMaker(makerName string) (*MarketMaker, error) {
	if maker, ok := cp.makers[makerName]; ok {
		return maker, nil
	}
	return nil, fmt.Errorf("no maker")
}

func (cp *CoinPair) addMaker(makerName string, valueA, valueB float64) {
	if _, ok := cp.makers[makerName]; !ok {
		maker := &MarketMaker{
			CoinA: &Coin{
				Name: cp.CoinA.Name,
			},
			CoinB: &Coin{
				Name: cp.CoinB.Name,
			},
		}
		cp.makers[makerName] = maker
	}

	cp.makers[makerName].CoinA.Add(valueA)
	cp.makers[makerName].CoinB.Add(valueB)
}

func (cp *CoinPair) getCoin(name string) (*Coin, error) {
	switch name {
	case cp.CoinA.Name:
		return cp.CoinA, nil
	case cp.CoinB.Name:
		return cp.CoinB, nil
	default:
		return nil, fmt.Errorf("failed to get coin: illegal coin name")
	}
}

func (cp *CoinPair) checkk() error {
	if cp.CoinB.Amount*cp.CoinA.Amount != cp.totalK {
		return fmt.Errorf("invalid k")
	}
	return nil
}

func (cp *CoinPair) updatek() {
	cp.totalK = cp.CoinB.Amount * cp.CoinA.Amount
}
