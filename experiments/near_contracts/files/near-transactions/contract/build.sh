#!/bin/bash

echo ">> Building contract"

near-sdk-js build src/uber.ts build/uber.wasm
near-sdk-js build src/fifa.ts build/fifa.wasm
near-sdk-js build src/exchange.ts build/exchange.wasm
near-sdk-js build src/youtube.ts build/youtube.wasm
near-sdk-js build src/twitter.ts build/dota.wasm
