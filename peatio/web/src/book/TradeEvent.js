class TradeEvent {
  constructor(book, side, price, executedQuantity) {
    this.book = book;
    this.side = side;
    this.price = price;

    this.executedQuantity = executedQuantity;
  }
}

export default TradeEvent;
