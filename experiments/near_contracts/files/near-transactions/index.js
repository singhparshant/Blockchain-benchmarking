const { connect, keyStores, Contract } = require('near-api-js');
const path = require('path');
const homedir = require('os').homedir();

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

async function get_testnet() {
  const config = {
    keyStore,
    networkId: 'testnet',
    nodeUrl: 'https://rpc.testnet.near.org',
  };
  return await connect(config);
}

async function interact_counter(account) {
  const contract = new Contract(account, 'test.near', {
    viewMethods: ['getCount'],
    changeMethods: ['increase', 'decrease'],
  });
  // console.log(
  //   await contract.account.functionCall('sipars.testnet', 'getCount'),
  // );
  // console.log(await account.functionCall('sipars.testnet', 'getCount'));
  console.log(
    await contract.getCount({ args: {}, gas: '300000', amount: '0' }),
  );
}

async function interact_twitter(account) {
  const contract = new Contract(account, 'sipars.testnet', {
    changeMethods: ['push', 'add', 'sub'],
    sender: account.accountId,
  });
  console.log(await contract.push({}));
}

async function interact_microsoft(account) {
  const contract = new Contract(account, 'sipars.testnet', {
    viewMethods: [],
    changeMethods: ['buy'],
  });
  console.log(
    await contract.buy({ args: {}, gas: '300000000000000', amount: '0' }),
  );
}

async function interact_uber(account) {
  const contract = new Contract(account, 'sipars.testnet', {
    viewMethods: ['checkDistance'],
    sender: account.accountId,
  });
  const gasLimit = '300000000000000'; // Adjust the value as needed

  console.log(await contract.checkDistance({}));
}

async function interact_youtube(account) {
  const contract = new Contract(account, 'sipars.testnet', {
    viewMethods: ['getModification'],
    callMethods: ['upload'],
    sender: account.accountId,
  });
  const gasLimit = '300000000000000'; // Adjust the value as needed

  console.log(await contract.getModification({ accountId: 'testId' }));
}

(async function () {
  // const near = await get_testnet();
  const near = await get_localnet();
  const account = await near.account('test.near');
  // console.log(account.deployContract());
  interact_counter(account);
  // interact_twitter(account);
  // interact_microsoft(account);
  // interact_uber(account);
  // interact_youtube(account);
})();
