package forms

import (
	"testing"

	"github.com/FlowerWrong/exchange/models"
	"github.com/devfeel/mapper"
)

func TestOrderBookFormMapper(t *testing.T) {
	obf := &OrderBookForm{
		Symbol:    "BTC_USD",
		OrderType: "limit",
		Side:      "Buy",
		Volume:    10.00,
		Price:     100.00,
	}
	ob := &models.OrderBook{}
	mapper.AutoMapper(obf, ob)
	t.Log(obf)
	t.Log(ob)
	if obf.Symbol != ob.Symbol {
		t.Fatal("Wrong done id")
	}
}
