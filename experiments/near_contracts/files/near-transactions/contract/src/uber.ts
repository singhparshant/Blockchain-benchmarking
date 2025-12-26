// Find all our documentation at https://docs.near.org
import { NearBindgen, call, Vector, view, near, initialize } from 'near-sdk-js';

@NearBindgen({})
class DecentralizedUber {
  number_of_drivers: number = 100;
  minDistance: number = 999999;

  driver_x: Vector<number> = new Vector<number>('unique-id-driver-x');
  driver_y: Vector<number> = new Vector<number>('unique-id-driver-y');

  @initialize({ privateFunction: true })
  init() {
    for (let i: number = 0; i < this.number_of_drivers; i++) {
      this.driver_x.push(Math.floor(Math.random() * 100));
      this.driver_y.push(Math.floor(Math.random() * 100));
    }
  }

  @call({})
  findClosestDriver(): number {
    let closer_driver: number;
    let d: number = this.minDistance;
    let client_x = Math.floor(Math.random() * 100);
    let client_y = Math.floor(Math.random() * 100);
    let diff_x: number;
    let diff_y: number;

    near.log(`driver_x 0 is ${this.driver_x.get(0)}`);

    for (let i: number = 0; i < this.number_of_drivers; i++) {
      // near.log(`minimum distance is ${this.driver_x.get(i)}`);
      diff_x =
        (client_x - (this.driver_x.get(i) || 0)) *
        (client_x - (this.driver_x.get(i) || 0));
      diff_y =
        (client_y - (this.driver_y.get(i) || 0)) *
        (client_y - (this.driver_y.get(i) || 0));

      let val = Math.sqrt(diff_x + diff_y);

      if (val < d) {
        d = val;
        closer_driver = i;
      }
    }
    near.log(`minimum distance is ${d}`);
    return closer_driver;
  }

  @view({})
  getPositions(): number {
    return this.driver_x.length + this.driver_y.length;
  }
}
