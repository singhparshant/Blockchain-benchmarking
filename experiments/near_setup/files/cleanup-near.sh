#!/bin/bash

# Check if the first script argument is present
if [ -z "$1" ]; then
    echo "No node name supplied. Exiting."
    exit 1
fi

# Kill all processes with the 'neard' binary running
pkill -f neard

# Clear everything in the directory ~/.near/
rm -rf ~/.near*

# Clear the results folder
rm -rf ~/results