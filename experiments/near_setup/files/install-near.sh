#!/bin/bash

set -e

curl -L https://nextcloud.in.tum.de/index.php/s/GBBMzL98dsZaRnt/download/near.tar.gz | tar -xzf -

# Download the Node.js 16.9.0 binary
curl -O https://nodejs.org/dist/v16.9.0/node-v16.9.0-linux-x64.tar.xz

# Extract the downloaded archive
tar -xf node-v16.9.0-linux-x64.tar.xz

# Copy to /usr/local
sudo cp -R node-v16.9.0-linux-x64/{bin,include,lib,share} /usr/local/

# sudo apt-get install -y npm=8.19.3
npm i -g pnpm@6.35.1
pnpm i -g near-sdk-js
npm i -g -y near-cli@3.4.2

exit 0
