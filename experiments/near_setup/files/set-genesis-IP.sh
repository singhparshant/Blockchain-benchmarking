#!/bin/bash

node=$1
folder=$2

cd $folder

node_number=$(echo $node | grep -o -E '[0-9]+')
let node_number-=1

ip=$(awk -v line=$node_number 'NR == line { print $1; exit }' genesis_ips.txt)
echo $ip
