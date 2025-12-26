import {
  NearBindgen,
  call,
  view,
  LookupMap,
  initialize,
  near,
} from 'near-sdk-js';

@NearBindgen({})
export class DecentralizedExchange {
  stocks: LookupMap<number> = new LookupMap<number>('stock');

  @initialize({ privateFunction: true })
  init() {
    this.stocks.set('MSFT', 10000000);
    this.stocks.set('AMZN', 10000000);
    this.stocks.set('GOOGL', 10000000);
    this.stocks.set('AAPL', 10000000);
    this.stocks.set('FB', 10000000);
  }

  @call({})
  buy({ stock }: { stock: string }): void {
    this.checkStock(stock, 1);
    let val = this.stocks.get(stock) || 0;
    if (val) this.stocks.set(stock, val - 1);
  }

  private checkStock(stock: string, value: number): void {
    if (this.stocks.get(stock) < value) {
      throw 'Not enough stocks';
    } else {
      near.log(`Number of stocks left: ${this.stocks.get(stock)}`);
    }
  }

  @view({})
  getStock({ stock }: { stock: string }): number {
    return this.stocks.get(stock) || 0;
  }
}
