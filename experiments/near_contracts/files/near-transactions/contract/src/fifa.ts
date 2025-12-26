import { NearBindgen, near, call, view, initialize } from 'near-sdk-js';

@NearBindgen({})
export class DecentralizedFifa {
  count = 0;

  @initialize({ privateFunction: true })
  init() {}

  @call({})
  increase() {
    this.count += 1;
    near.log(`fifa counter increased to ${this.count}`);
  }

  @call({})
  decrease() {
    this.count -= 1;
    near.log(`fifa counter decreased to ${this.count}`);
  }

  @view({})
  getCount(): number {
    return this.count;
  }
}
