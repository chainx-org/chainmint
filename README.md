# chainmint
chainmint = (tendermint + utxo + cvm )。 类似于ethermint， chain.com是基于tendermint 实现的abci应用。


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
