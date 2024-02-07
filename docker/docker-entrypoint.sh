#!/usr/bin/env sh
set -e

DATA_DIR=${DATA_DIR:="/data"}
CONFIG_DIR=${DATA_DIR}/config
JWT_SECRET=${JWT_SECRET:=$(openssl rand -hex 32 )}
GENESIS_FILE=$CONFIG_DIR/genesis.json
OP_GETH_GENESIS_URL=https://networks.optimism.io/op-sepolia/genesis.json

if [ ! -s $GENESIS_FILE ]; then
  mkdir -p $CONFIG_DIR
  curl $OP_GETH_GENESIS_URL -o $GENESIS_FILE
  geth init --datadir $DATA_DIR $GENESIS_FILE
else
  echo "Genesis file already exists. Skipping initialization."
fi

echo $JWT_SECRET > $CONFIG_DIR/jwtsecret

supervisord -n -c  /etc/supervisor/conf.d/supervisord.conf
