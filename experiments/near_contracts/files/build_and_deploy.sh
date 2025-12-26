#!/bin/bash

# Install dependencies
cd ~/near-transactions && pnpm install
cd ~/near-transactions/contract &&  pnpm install

# Build smart contracts
cd ~/near-transactions/contract
# SCRIPT_PATH="./build.sh"
# "$SCRIPT_PATH"

# Deploy smart contracts
chown -R $USER:$USER /root
npm run deploy