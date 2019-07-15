package models

import (
	"strconv"

	"github.com/FlowerWrong/exchange/db"
	"github.com/FlowerWrong/exchange/services/matching"
	"github.com/shopspring/decimal"
)

// OrderBook ...
type OrderBook struct {
	BaseModel
	UserID    uint64          `json:"user_id"`
	User      User            `json:"-"`
	Symbol    string          `json:"symbol"`
	FundID    uint64          `json:"fund_id"`
	Fund      Fund            `json:"-"`
	Status    uint            `json:"status"`     // pending done cancel reject
	OrderType string          `json:"order_type"` // market or limit
	Side      string          `json:"side"`       // sell or buy
	Volume    decimal.Decimal `json:"volume" sql:"DECIMAL(32,16)"`
	Price     decimal.Decimal `json:"price" sql:"DECIMAL(32,16)"`
}

// OtherSide sell -> buy; buy -> sell
func (ob *OrderBook) OtherSide() string {
	if ob.Side == "sell" {
		return "buy"
	}
	return "sell"
}

// StrID return string id
func (ob *OrderBook) StrID() string {
	return strconv.FormatUint(ob.ID, 10)
}

// Transaction ...
func Transaction(orderBook *OrderBook, done []*matching.Order) error {
	tx := db.ORM().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	tx.Create(orderBook)

	// 账户余额锁定
	account := &Account{}
	fund := &Fund{}
	tx.First(fund, orderBook.FundID)
	if orderBook.Side == "buy" {
		// BTC_USD 为例，购买动作即用USD买BTC，锁定账户的USD
		FindAccountByUserIDAndCurrencyID(tx, account, orderBook.UserID, fund.RightCurrencyID)
		account.Lock(orderBook.Volume.Mul(orderBook.Price)) // 单价 * 数量
	} else {
		FindAccountByUserIDAndCurrencyID(tx, account, orderBook.UserID, fund.LeftCurrencyID)
		account.Lock(orderBook.Volume)
	}
	tx.Save(account)

	for _, matchingOrderDone := range done {
		id := matchingOrderDone.IntID()

		// 对方记录
		orderBookDone := &OrderBook{}
		tx.Find(orderBookDone, id)
		orderBookDone.Volume = orderBookDone.Volume.Sub(matchingOrderDone.Quantity())
		if orderBookDone.Volume.Sign() == 0 {
			orderBookDone.Status = 1
		}
		tx.Save(orderBookDone)

		// 保存交易记录
		orderOther := &Order{}
		orderOther.OrderBookID = id
		orderOther.UserID = orderBookDone.UserID
		orderOther.FundID = orderBookDone.FundID
		orderOther.Symbol = orderBookDone.Symbol
		orderOther.OrderType = orderBookDone.OrderType
		orderOther.Side = orderBookDone.Side
		orderOther.Volume = matchingOrderDone.Quantity()
		orderOther.Price = matchingOrderDone.Price()
		tx.Create(orderOther)

		// 账户结算
		Settlement(orderOther, tx)

		// 当前用户记录
		orderBook.Volume = orderBook.Volume.Sub(matchingOrderDone.Quantity())
		if orderBook.Volume.Sign() == 0 {
			orderBook.Status = 1
		}
		tx.Save(orderBook)

		// 保存交易记录
		order := &Order{}
		order.OrderBookID = orderBook.ID
		order.UserID = orderBook.UserID
		order.FundID = orderBook.FundID
		order.OtherSideOrderBookID = orderBookDone.ID
		order.OtherSideOrderID = orderOther.ID
		order.Symbol = orderBook.Symbol
		order.OrderType = orderBook.OrderType
		order.Side = orderBook.Side
		order.Volume = matchingOrderDone.Quantity()
		order.Price = matchingOrderDone.Price()
		tx.Create(order)

		// 账户结算
		Settlement(order, tx)

		orderOther.OtherSideOrderBookID = orderBook.ID
		orderOther.OtherSideOrderID = order.ID
		tx.Save(orderOther)
	}

	if orderBook.OrderType == "market" {
		orderBook.Status = 1
		tx.Save(orderBook)
	}
	return tx.Commit().Error
}
