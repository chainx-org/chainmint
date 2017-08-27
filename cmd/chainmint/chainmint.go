package main

import (
	"fmt"
	"os"
//	"strings"
//	"gopkg.in/urfave/cli.v1"

	abciApp "github.com/chainmint/app"
	//cmtUtils "github.com/chainmint/cmd/utils"
//	"github.com/chainmint/core"
	"github.com/chainmint/chain"
	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
)

const (
	addr = "tcp://127.0.0.1:46658"
	abci = "grpc"
)

func chainmintCmd(/*ctx *cli.Context*/) error {
	// Setup the ABCI server and start it
//	addr := ctx.GlobalString(cmtUtils.ABCIAddrFlag.Name)
//	abci := ctx.GlobalString(cmtUtils.ABCIProtocolFlag.Name)
	// Fetch the registered service of this type
	//rpcClient, err := node.Attach()
	//if err != nil {
	//	ethUtils.Fatalf("Failed to attach to the inproc geth: %v", err)
	//}

	// Create the ABCI app
	chainApp := abciApp.NewChainmintApplication( nil)
	// Start the app on the ABCI server
	chain.Run(chainApp)
	srv, err := server.NewServer(addr, abci, chainApp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//srv.SetLogger(cmtUtils.GetTMLogger().With("module", "chainmint"))
	if _, err := srv.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cmn.TrapSignal(func() {
		srv.Stop()
	})
	return nil
}
