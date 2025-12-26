// Find all our documentation at https://docs.near.org
import { NearBindgen, call, Vector, view, near, initialize } from 'near-sdk-js';

@NearBindgen({})
class DecentralizedDota {
  number_of_players = 10;
  maxlen = 250;

  team_positions_x: Vector<number> = new Vector<number>('x_team_positions');
  team_positions_y: Vector<number> = new Vector<number>('y_team_positions');

  @initialize({ privateFunction: true })
  init() {
    for (let i: number = 0; i < this.number_of_players; i++) {
      this.team_positions_x.push(Math.floor(Math.random() * 250));
      this.team_positions_y.push(Math.floor(Math.random() * 250));
    }
  }

  @call({})
  update(): void {
    for (let i: number = 0; i < this.number_of_players; i++) {
      this.team_positions_x.replace(i, (this.team_positions_x.get(i) || 0) + 1);
      this.team_positions_y.replace(i, (this.team_positions_y.get(i) || 0) + 1);

      if ((this.team_positions_x.get(i) || 0) > this.maxlen) {
        this.team_positions_x.replace(i, 0);
      }

      if ((this.team_positions_x.get(i) || 0) < 0) {
        this.team_positions_x.replace(i, this.maxlen);
      }

      if ((this.team_positions_y.get(i) || 0) > this.maxlen) {
        this.team_positions_y.replace(i, 0);
      }

      if ((this.team_positions_y.get(i) || 0) < 0) {
        this.team_positions_y.replace(i, this.maxlen);
      }
    }
    near.log(
      'Updated positions',
      this.team_positions_x.get(1),
      this.team_positions_y.get(2),
    );
  }

  @view({})
  getPositions(): number {
    return this.team_positions_x.length + this.team_positions_y.length;
  }
}
