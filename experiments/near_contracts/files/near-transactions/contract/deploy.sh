#!/bin/bash

echo ">> Deploying contract"

# https://docs.near.org/tools/near-cli#near-dev-deploy
contracts=("fifa" "youtube" "uber" "exchange" "dota")

# Loop over the nodes and for each of them run the near deploy command
for contract in "${contracts[@]}"; do
  # These accounts are for contract deployment
  # We need different accounts because near doesn't allow multiple smart contracts from a single account
  NEAR_ENV=localnet near --keyPath ~/.near/node-1/validator_key.json create-account $contract.node-1 --masterAccount node-1 --initialBalance 100
  echo "y" | NEAR_ENV=localnet near deploy --wasmFile build/$contract.wasm --initFunction init --initArgs '{}' --accountId $contract.node-1 --initGas=300000000000000
done
