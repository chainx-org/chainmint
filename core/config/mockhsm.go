//+build !no_mockhsm

package config

import (
	"context"

	"github.com/chainmint/core/mockhsm"
	"github.com/chainmint/crypto/ed25519"
	"github.com/chainmint/database/pg"
	"github.com/chainmint/log"
)

func getOrCreateDevKey(ctx context.Context, db pg.DB, c *Config) (blockPub ed25519.PublicKey, err error) {
	hsm := mockhsm.New(db)
	corePub, created, err := hsm.GetOrCreate(ctx, autoBlockKeyAlias)
	if err != nil {
		return nil, err
	}
	if created {
		log.Printf(ctx, "Generated new block-signing key %s\n", corePub.Pub)
	} else {
		log.Printf(ctx, "Using block-signing key %s\n", corePub.Pub)
	}
	c.BlockPub = corePub.Pub

	return corePub.Pub, nil

}

func checkBlockHSMURL(string) error {
	return nil
}
