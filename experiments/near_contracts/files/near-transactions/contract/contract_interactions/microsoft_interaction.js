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

async function interact_microsoft(account) {
  const contract = new Contract(account, 'microsoft.test.near', {
    viewMethods: [],
    changeMethods: ['buy'],
  });
  const result = [];
  for (let i = 0; i < 5; i++) {
       result.push(await contract.buy({ args: {}, gas: '300000000000000', amount: '0' }),)
  }
  fs.writeFile('counter.csv', result.join(','), function (err) {
    if (err) return console.log(err);
});
}


(async function () {
  const near = await get_localnet();
  const account = await near.account('microsoft.test.near');
  interact_microsoft(account);
})();
