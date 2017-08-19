package types

import (
	"github.com/tendermint/abci/types"
	"github.com/chainmint/protocol/bc/legacy"
)

// ValidatorsStrategy is a validator strategy
type ValidatorsStrategy interface {
	SetValidators(validators []*types.Validator)
	CollectTx(tx *legacy.Tx)
	GetUpdatedValidators() []*types.Validator
}

// Strategy encompasses all available strategies
type Strategy struct {
	ValidatorsStrategy
}
