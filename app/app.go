package app

import (
	"encoding/json"
	"context"
//	"fmt"
//	"math/big"

	"github.com/chainmint/protocol/state"
	//"github.com/chainmint/protocol"
	"github.com/chainmint/core"
	"github.com/chainmint/protocol/bc/legacy"
	"github.com/chainmint/log"
	//"github.com/chainmint/protocol/bc"
	abciTypes "github.com/tendermint/abci/types"

	cmtTypes "github.com/chainmint/types"
)

// ChainmintApplication implements an ABCI application
type ChainmintApplication struct {

	// backend handles the chain state machine
	// and wrangles other services started by an chain node (eg. tx pool)
	backend *core.API // backend chain struct

	// a closure to return the latest current state from the chain
	currentState func() (*legacy.Block, *state.Snapshot)

	// strategy for validator compensation
	strategy *cmtTypes.Strategy
}

// NewChainmintApplication creates the abci application for Chainmint
func NewChainmintApplication(strategy *cmtTypes.Strategy) *ChainmintApplication {
	app := &ChainmintApplication{
		strategy:     strategy,
	}
	return app
}

func (app *ChainmintApplication) Init(backend *core.API/*, client *rpc.Client*/) {
	app.backend = backend
	app.currentState = backend.Chain().State
}

// Info returns information about the last height and app_hash to the tendermint engine
func (app *ChainmintApplication) Info() abciTypes.ResponseInfo {
	log.Printf(context.Background(), "Info")
	currentBlock, _ := app.currentState()
	if currentBlock == nil {
		return abciTypes.ResponseInfo{
			Data:   "ABCIChain",
			LastBlockHeight: uint64(0),
			LastBlockAppHash: []byte{},
		}
	}
	height := currentBlock.BlockHeight()
	hash := currentBlock.BlockHash()
	/*blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()
	*/

	// This check determines whether it is the first time chainmint gets started.
	// If it is the first time, then we have to respond with an empty hash, since
	// that is what tendermint expects.
	if height == 0 {
		return abciTypes.ResponseInfo{
			Data:             "ABCIChain",
			LastBlockHeight:  uint64(0),
			LastBlockAppHash: []byte{},
		}
	}

	return abciTypes.ResponseInfo{
		Data:             "ABCIChain",
		LastBlockHeight:  height,
		LastBlockAppHash: hash,
	}
}

// SetOption sets a configuration option
func (app *ChainmintApplication) SetOption(key string, value string) (log string) {
	//log.Info("SetOption")
	return ""
}

// InitChain initializes the validator set
func (app *ChainmintApplication) InitChain(validators []*abciTypes.Validator) {
	log.Printf(context.Background(), "InitChain")
	//app.setvalidators(validators)
	app.SetValidators(validators)
}

// CheckTx checks a transaction is valid but does not mutate the state
func (app *ChainmintApplication) CheckTx(txBytes []byte) abciTypes.Result {
	tx, err := decodeTx(txBytes)
	log.Printf(context.Background(), "Received CheckTx", "tx", tx)
	if err != nil {
		return abciTypes.ErrEncodingError.AppendLog(err.Error())
	}

	return app.validateTx(tx)
}

// DeliverTx executes a transaction against the latest state
func (app *ChainmintApplication) DeliverTx(txBytes []byte) abciTypes.Result {
	tx, err := decodeTx(txBytes)
	if err != nil {
		return abciTypes.ErrEncodingError.AppendLog(err.Error())
	}

	log.Printf(context.Background(), "Got DeliverTx", "tx", tx)
	app.backend.Generator().Submit(context.Background(), tx)
	app.CollectTx(tx)

	return abciTypes.OK
}

// BeginBlock starts a new chain block
func (app *ChainmintApplication) BeginBlock(hash []byte, tmHeader *abciTypes.Header) {
	log.Printf(context.Background(), "BeginBlock")
}

// EndBlock accumulates rewards for the validators and updates them
func (app *ChainmintApplication) EndBlock(height uint64) abciTypes.ResponseEndBlock {
	log.Printf(context.Background(), "EndBlock")
	return app.GetUpdatedValidators()
}

// Commit commits the block and returns a hash of the current state
func (app *ChainmintApplication) Commit() abciTypes.Result {
	log.Printf(context.Background(), "Commit")
	app.backend.Generator().MakeBlock(context.Background())
	blockHash := []byte("")
	return abciTypes.NewResultOK(blockHash[:], "")
}

// Query queries the state of ChainmintApplication
func (app *ChainmintApplication) Query(query abciTypes.RequestQuery) abciTypes.ResponseQuery {
	log.Printf(context.Background(), "Query")
	/*var in jsonRequest
	if err := json.Unmarshal(query.Data, &in); err != nil {
		return abciTypes.ResponseQuery{Code: abciTypes.ErrEncodingError.Code, Log: err.Error()}
	}*/
	var result interface{}
	/*if err := app.rpcClient.Call(&result, in.Method, in.Params...); err != nil {
		return abciTypes.ResponseQuery{Code: abciTypes.ErrInternalError.Code, Log: err.Error()}
	}*/

	bytes, _ := json.Marshal(result)
	bytes = []byte("")
/*	bytes, err := json.Marshal(result)
	if err != nil {
		return abciTypes.ResponseQuery{Code: abciTypes.ErrInternalError.Code, Log: err.Error()}
	}*/
	return abciTypes.ResponseQuery{Code: abciTypes.OK.Code, Value: bytes}
}

//-------------------------------------------------------

// validateTx checks the validity of a tx against the blockchain's current state.
// it duplicates the logic in ethereum's tx_pool
func (app *ChainmintApplication) validateTx(tx *legacy.Tx) abciTypes.Result {
	/*currentState, err := app.currentState()
	if err != nil {
		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return abciTypes.ErrBaseInvalidSignature.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Make sure the account exist. Non existent accounts
	// haven't got funds and well therefor never pass.
	if !currentState.Exist(from) {
		return abciTypes.ErrBaseUnknownAddress.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Check for nonce errors
	currentNonce := currentState.GetNonce(from)
	if currentNonce > tx.Nonce() {
		return abciTypes.ErrBadNonce.
			AppendLog(fmt.Sprintf("Got: %d, Current: %d", tx.Nonce(), currentNonce))
	}

	// Check the transaction doesn't exceed the current block limit gas.
	gasLimit := app.backend.GasLimit()
	if gasLimit.Cmp(tx.Gas()) < 0 {
		return abciTypes.ErrInternalError.AppendLog(core.ErrGasLimitReached.Error())
	}

	// Transactions can't be negative. This may never happen
	// using RLP decoded transactions but may occur if you create
	// a transaction using the RPC for example.
	if tx.Value().Cmp(common.Big0) < 0 {
		return abciTypes.ErrBaseInvalidInput.
			SetLog(core.ErrNegativeValue.Error())
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	currentBalance := currentState.GetBalance(from)
	if currentBalance.Cmp(tx.Cost()) < 0 {
		return abciTypes.ErrInsufficientFunds.
			AppendLog(fmt.Sprintf("Current balance: %s, tx cost: %s", currentBalance, tx.Cost()))

	}

	intrGas := core.IntrinsicGas(tx.Data(), tx.To() == nil, true) // homestead == true
	if tx.Gas().Cmp(intrGas) < 0 {
		return abciTypes.ErrBaseInsufficientFees.
			SetLog(core.ErrIntrinsicGas.Error())
	}
*/
	return abciTypes.OK
}
