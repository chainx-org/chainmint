package app

import (
	"bytes"
	"encoding/json"
	github.com/chain/encoding/blockchain
	github.com/chain/protocol/bc/legacy

	abciTypes "github.com/tendermint/abci/types"
)

// format of query data
type jsonRequest struct {
	Method string          `json:"method"`
	ID     json.RawMessage `json:"id,omitempty"`
	Params []interface{}   `json:"params,omitempty"`
}

// decode to chain's transaction
func decodeTx(txBytes []byte) (*legacy.Tx, error) {
	var tx legacy.Tx
	err := UnmarshalText(txBytes)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

//-------------------------------------------------------
// convenience methods for validators

// Receiver returns the receiving address based on the selected strategy
func (app *ChainmintApplication) Receiver() common.Address {
	if app.strategy != nil {
		return app.strategy.Receiver()
	}
	return common.Address{}
}

// SetValidators sets new validators on the strategy
func (app *ChainmintApplication) SetValidators(validators []*abciTypes.Validator) {
	if app.strategy != nil {
		app.strategy.SetValidators(validators)
	}
}

// GetUpdatedValidators returns an updated validator set from the strategy
func (app *ChainmintApplication) GetUpdatedValidators() abciTypes.ResponseEndBlock {
	if app.strategy != nil {
		return abciTypes.ResponseEndBlock{Diffs: app.strategy.GetUpdatedValidators()}
	}
	return abciTypes.ResponseEndBlock{}
}

// CollectTx invokes CollectTx on the strategy
func (app *ChainmintApplication) CollectTx(tx *legacy.Tx) {
	if app.strategy != nil {
		app.strategy.CollectTx(tx)
	}
}
