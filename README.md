# [Chainmint](https://github.com/chainx-org/chainmint)  **=** [Tendermint](https://tendermint.com/) **+** [Chain](https://chain.com/)

Chainmint is based on the tendermint consensus inherited from Chain's UTXO and CVM. It can become a Cosmos Zone in the future, supporting [Chain](https://chain.com) cross-chain functionality. For simplicity:

<p align="center">
<b>Chainmint = Tendermint + UTXO + CVM</b>
</p>


## Components

- [**Chainmint**](https://github.com/chainx-org/chainmint): implements the specific logic of the abci interface.
- [**Tendermint**](https://github.com/tendermint/tendermint): consensus module, handles chainmint transaction order.
- [**PostgreSql**](https://github.com/postgres/postgres): data storage module.
- [**Chainmintcli**](https://github.com/chainx-org/chainmint/tree/master/cmd/chainmintcli): client for communication with chainmint.

## Getting Started

### Prerequisites

1. install tendermint
2. install postgreSQL

### Build

``` bash
make get_vendor_deps
cd cmd/chainmint
go build
cd cmd/chainmintcli
go build
```

### Run

#### Chainmint

First, configure `user`, `password`, `dbname` and `sslmode` in `chainmint/chain/run.go`:

``` go
dbURL = env.String("DATABASE_URL", "user=yourusername password=yourpassword dbname=core sslmode=disable")
```

then execute `chainmint/core/schema.sql` in postgreSql's `core` (i.e., `dbname`) database.

Enter `chainmint/cmd/chainmint` and run `./chainmint`.

#### Local Tendermint

``` bash
./tendermint init --home ./yourdir
./tendermint node --home ./yourdir
```

#### Chainmintcli
``` bash
./chainmintcli <options.>
```

more details: [chainmint(Chinese)](https://gguoss.github.io/2017/09/03/chainmint/)
