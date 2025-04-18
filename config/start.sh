#!/bin/sh

set -e

biome="Plains"
groundlevel="62"
sealevel="0"
finishers=""

if [ -n "$1" ]; then
    biome="$1"
fi

if [ -n "$2" ]; then
    groundlevel="$2"
fi

if [ -n "$3" ]; then
    sealevel="$3"
fi

if [ -n "$4" ]; then
    finishers="$4"
fi

sed -i "s/@BIOME@/${biome}/g;s/@GROUNDLEVEL@/${groundlevel}/g;s/@SEALEVEL@/${sealevel}/g;s/@FINISHERS@/${finishers}/g" /opt/Server/world/world.ini

# Start goproxy
#goproxy > /opt/Server/world/goproxy_out 2>&1 &

kubeproxy 2>&1 | tee /opt/Server/world/kubeproxy_out 2>&1 &

# start Minecraft C++ server
cd /opt/Server
./Cuberite
