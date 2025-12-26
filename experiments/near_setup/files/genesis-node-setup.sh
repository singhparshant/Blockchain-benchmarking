#!/bin/bash

# Redirect all output to ~/genesis-node-setup-output.txt
exec > ~/genesis-node-setup-output.txt 2>&1

# Get arguments
num_validators=$1
num_non_validators=$2
shards=$3
total_nodes=$((num_validators + num_non_validators))

# Initialize the localnet
cd ~/nearcore/target/release
./neard --home ~/.near localnet --validators ${num_validators} --non-validators ${num_non_validators} --shards ${shards} --prefix node-

# Start the numbering of nodes from 1. By default it starts from 0
for ((i=total_nodes-1; i>=0; i--)); do
    old_name="node-$i"
    new_name="node-$(($i+1))"
    mv "$HOME/.near/$old_name" "$HOME/.near/$new_name"
done

# Loop through each node and modify the files
for ((i=1; i<=total_nodes; i++)); do
    node_name="node-$i"
    config_file="$HOME/.near/${node_name}/config.json"
    genesis_file="$HOME/.near/${node_name}/genesis.json"
    validator_key_file="$HOME/.near/${node_name}/validator_key.json"

    # Modify config.json
    if [[ -f $config_file ]]; then
        sed -i 's/"tracked_shards": \[\]/"tracked_shards": [0]/' "$config_file"
        echo "Modified $config_file"
    else
        echo "File not found: $config_file"
    fi

    # # Modify genesis.json to enable dynamic resharding
    # if [[ -f $genesis_file ]]; then
    #     sed -i 's/"dynamic_resharding": false/"dynamic_resharding": true/' "$genesis_file"
    #     echo "Enabled dynamic resharding in $genesis_file"
    # else
    #     echo "File not found: $genesis_file"
    # fi

    # Modify node numbers in genesis.json and validator_key.json
    for file in "$genesis_file" "$validator_key_file"; do
        if [[ -f $file ]]; then
            awk -F '"' '/"account_id": "node-/ {split($4,a,"-"); $4="node-"a[2]+1}1' OFS='"' "$file" > tmp && mv tmp "$file"
            echo "Modified node numbers in $file"
        else
            echo "File not found: $file"
        fi
    done
done
