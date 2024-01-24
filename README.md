# Blockchain Listener

Blockchain Listener is a service that will listen for blockchain events and publish them to queue.

## Quick Start

Clone code into local machine:

```sh
git clone git@github.com:KyberNetwork/evmlistener.git
cd evmlistener
```

Create environment file with following content, `listener.env`:

```sh
export WS_RPC="wss://polygon.kyberengineering.io"
export HTTP_RPC="https://polygon.kyberengineering.io"
export SANITY_NODE_RPC="https://polygon.kyberengineering.io"
export SANITY_CHECK_INTERVAL=10s
export LOG_LEVEL="debug"

export SENTRY_DNS=""
export SENTRY_LEVEL="error"

export REDIS_MASTER_NAME=""
export REDIS_ADDRS="localhost:6379"
export REDIS_DB=0
export REDIS_USERNAME=""
export REIDS_PASSWORD=""
export REDIS_KEY_PREFIX="test-listener-polygon:"
export REDIS_READ_TIMEOUT=0
export REIDS_WRITE_TIMEOUT=0

export PUBLISHER_TOPIC="test-listener-polygon-topic"
export PUBLISHER_MAX_LEN=10

export MAX_NUM_BLOCKS=128
export BLOCK_EXPIRATION=10m
```

Start docker for redis:

```sh
docker-compose up -d
```

Run service:

```sh
source listener.env
go run ./cmd/listener/main.go
```

## Other

To re-generate proto files:
```
protoc --go_out protobuf protobuf/ethereum.proto protobuf/message.proto
```
