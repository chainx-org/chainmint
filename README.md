# [chainmint](https://github.com/chainx-org/chainmint)  **=** [Tendermint](https://tendermint.com/) **+** [Chain](https://chain.com/) 
chainmint = (tendermint + utxo + cvm ).
chainmint is based on the tendermint consensus, inherited Chain's UTXO, CVM block chain. The future will become cosmos a zone, used to support Chain (Nasdaq stock market block technology used) cross-chain technology.

## components
- **Tendermint** (https://github.com/tendermint/tendermint) chainmint consensus module, organized chainmint transaction order.
- **Chainmint** (https://github.com/chainx-org/chainmint) implements the specific logic of the abci interface.
- **PostgreSql** (https://github.com/postgres/postgres) chainmint data storage module.
- **Chainmintcli** (https://github.com/chainx-org/chainmint/tree/master/cmd/chainmintcli) chainmint client for communication with chainmint.

## build steps
``` console
make get_vendor_deps
cd cmd/chainmint
go build
cd cmd/chainmintcli
go build
```
## install steps
- 1. install tendermint
- 2. install postgreSQL

## run steps
### run chainmint
``` console
1. set dbURL         = env.String("DATABASE_URL", "user=yourusername password=yourpassword dbname=core sslmode=disable") in chainmint/chain/run.go
2. execute chainmint/core/schema.sql in postgreSql's core database.
3. ./chainmint
```
### run local tendermint
``` console
4. ./tendermint init --home ./yourdir
5. ./tendermint node --home ./yourdir
```
### run chainmintcli to test
``` console
6. ./chainmintcli <options.>
```

## more details:https://gguoss.github.io/2017/09/03/chainmint/
