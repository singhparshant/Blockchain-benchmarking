import { connect, keyStores, Contract } from 'near-api-js';
import path from 'path';
import fs from 'fs';
import os from 'os';

const homedir = os.homedir();
const CREDENTIALS_DIR = '.near-credentials';
const credentialsPath = path.join(homedir, CREDENTIALS_DIR);
const keyStore = new keyStores.UnencryptedFileSystemKeyStore(credentialsPath);

async function get_localnet() {
  const connectionConfig = {
    keyStore,
    networkId: 'local',
    nodeUrl: 'http://localhost:3030',
  };
  return await connect(connectionConfig);
}
async function interact_counter(account) {
  const contract = new Contract(account, 'counter.test.near', {
    viewMethods: ['getCount'],
    changeMethods: ['increase', 'decrease'],
  });
  // console.log(
  //   await contract.account.functionCall('sipars.testnet', 'getCount'),
  // );
  // console.log(await account.functionCall('sipars.testnet', 'getCount'));

  const result = [];
    for (let i = 0; i < 5; i++) {
      result.push(console.log(
        await contract.increase({
          args: {},
          gas: '30000000000000',
          amount: '0',
        })))
    }
    fs.writeFile('counter.csv', result.join(','), function (err) {
      if (err) return console.log(err);
  });
}

(async function () {
  const near = await get_localnet();
  const account = await near.account('counter.test.near');
  interact_counter(account);
})();
