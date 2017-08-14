package app

import (
	"bytes"
	"encoding/json"

	abciTypes "github.com/tendermint/abci/types"
)

// format of query data
type jsonRequest struct {
	Method string          `json:"method"`
	ID     json.RawMessage `json:"id,omitempty"`
	Params []interface{}   `json:"params,omitempty"`
}

// rlp decode an etherum transaction
func decodeTx(txBytes []byte) (*types.Transaction, error) {
	tx := new(types.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(txBytes), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
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
func (app *ChainmintApplication) CollectTx(tx *types.Transaction) {
	if app.strategy != nil {
		app.strategy.CollectTx(tx)
	}
}
