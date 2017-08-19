package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chainmint/core/account"
	"github.com/chainmint/core/asset"
	"github.com/chainmint/core/coretest"
	"github.com/chainmint/core/generator"
	"github.com/chainmint/core/pin"
	"github.com/chainmint/core/txbuilder"
	"github.com/chainmint/database/pg/pgtest"
	"github.com/chainmint/protocol/bc"
	"github.com/chainmint/protocol/prottest"
	"github.com/chainmint/testutil"
)

func BenchmarkBuildTx(b *testing.B) {
	b.StopTimer()

	cases := []struct {
		numAssets, numIssuances int
	}{
		{1, 1}, {1, 10}, {1, 100},
		{10, 1}, {10, 10}, {10, 100},
		{100, 1}, {100, 10},
	}

	_, db := pgtest.NewDB(b, pgtest.SchemaPath)
	ctx := context.Background()
	pinStore := pin.NewStore(db)
	chain := prottest.NewChain(b)
	accounts := account.NewManager(db, chain, pinStore)
	assets := asset.NewRegistry(db, chain, pinStore)
	accountID := coretest.CreateAccount(ctx, b, accounts, "account", nil)
	generator := generator.New(chain, nil, db)

	var assetIDs []bc.AssetID
	for _, c := range cases {
		if c.numAssets > len(assetIDs) {
			for i := len(assetIDs); i < c.numAssets; i++ {
				assetID := coretest.CreateAsset(ctx, b, assets, nil, fmt.Sprintf("asset%d", i), nil)
				assetIDs = append(assetIDs, assetID)
			}
		}
	}

	prepareActions := func(numAssets, numIssuances int) []txbuilder.Action {
		var actions []txbuilder.Action
		for i := 0; i < numAssets; i++ {
			assetID := assetIDs[i]
			for j := 0; j < numIssuances; j++ {
				actions = append(actions, assets.NewIssueAction(bc.AssetAmount{AssetId: &assetID, Amount: 1}, nil))
			}
			actions = append(actions, accounts.NewControlAction(bc.AssetAmount{AssetId: &assetID, Amount: uint64(numIssuances)}, accountID, nil))
		}
		return actions
	}

	doBuild := func(actions []txbuilder.Action) *txbuilder.Template {
		tpl, err := txbuilder.Build(ctx, nil, actions, time.Now().Add(time.Minute))
		if err != nil {
			b.Fatal(err)
		}
		return tpl
	}

	for _, c := range cases {
		actions := prepareActions(c.numAssets, c.numIssuances)
		name := fmt.Sprintf("%d-asset--%d-issuance", c.numAssets, c.numIssuances)
		b.Run(name+"--build", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				doBuild(actions)
			}
		})
		b.Run(name+"--build-sign", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tpl := doBuild(actions)
				coretest.SignTxTemplate(b, ctx, tpl, &testutil.TestXPrv)
			}
		})
		b.Run(name+"--build-sign-finalize", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tpl := doBuild(actions)
				coretest.SignTxTemplate(b, ctx, tpl, &testutil.TestXPrv)
				func() {
					dbtx := pgtest.NewTx(b)
					defer dbtx.Rollback(ctx)

					err := txbuilder.FinalizeTx(ctx, chain, generator, tpl.Transaction)
					if err != nil {
						b.Fatal(err)
					}
				}()
			}
		})
	}
}
