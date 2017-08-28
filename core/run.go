package core

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	//"errors"
	"net"
	"net/http"
	"time"

	"github.com/chainmint/core/accesstoken"
	"github.com/chainmint/core/account"
	"github.com/chainmint/core/asset"
	//"github.com/chainmint/core/config"
	//"github.com/chainmint/core/fetch"
	"github.com/chainmint/core/generator"
	//"github.com/chainmint/core/leader"
	"github.com/chainmint/core/pin"
	"github.com/chainmint/core/query"
	"github.com/chainmint/core/rpc"
	"github.com/chainmint/core/txbuilder"
	"github.com/chainmint/core/txdb"
	"github.com/chainmint/core/txfeed"
	"github.com/chainmint/database/pg"
	//"github.com/chainmint/database/raft"
	//"github.com/chainmint/log"
	"github.com/chainmint/protocol"
	"github.com/chainmint/protocol/bc/legacy"
	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
)

const (
	blockPeriod              = time.Second
	expireReservationsPeriod = time.Second
	tendermintLAddr          = "tcp://0.0.0.0:46657"
)

// RunOption describes a runtime configuration option.
type RunOption func(*API)

// UseTLS configures the Core to use TLS with the given config
// when communicating between Core processes.
// If c is nil, TLS is disabled.
func UseTLS(c *tls.Config) RunOption {
	return func(a *API) {
		a.useTLS = c != nil
		a.httpClient = new(http.Client)
		a.httpClient.Transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSClientConfig:       c,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		if c != nil {
			// TODO(kr): set Leaf in TLSConfig and use that here.
			x509Cert, err := x509.ParseCertificate(c.Certificates[0].Certificate[0])
			if err != nil {
				panic(err)
			}
			a.internalSubj = x509Cert.Subject
		}
	}
}

// BlockSigner configures the Core to use signFn to handle block-signing
// requests. In production, this will be a function to call out to signerd
// and its HSM. In development, it'll use the MockHSM.
func BlockSigner(signFn func(context.Context, *legacy.Block) ([]byte, error)) RunOption {
	return func(a *API) { a.signer = signFn }
}

// GeneratorLocal configures the launched Core to run as a Generator.
func GeneratorLocal(gen *generator.Generator) RunOption {
	return func(a *API) {
		if a.remoteGenerator != nil {
			panic("core configured with local and remote generator")
		}
		a.generator = gen
		a.submitter = gen
	}
}

// GeneratorRemote configures the launched Core to fetch blocks from
// the provided remote generator.
func GeneratorRemote(client *rpc.Client) RunOption {
	return func(a *API) {
		if a.generator != nil {
			panic("core configured with local and remote generator")
		}
		a.remoteGenerator = client
		a.submitter = &txbuilder.RemoteGenerator{Peer: client}
	}
}

// IndexTransactions configures whether or not transactions should be
// annotated and indexed for the query engine.
func IndexTransactions(b bool) RunOption {
	return func(a *API) { a.indexTxs = b }
}

// RateLimit adds a rate-limiting restriction, using keyFn to extract the
// key to rate limit on. It will allow up to burst requests in the bucket
// and will refill the bucket at perSecond tokens per second.
func RateLimit(keyFn func(*http.Request) string, burst, perSecond int) RunOption {
	return func(a *API) {
		a.requestLimits = append(a.requestLimits, requestLimit{
			key:       keyFn,
			burst:     burst,
			perSecond: perSecond,
		})
	}
}

// RunUnconfigured launches a new unconfigured Chain Core. This is
// used for Chain Core Developer Edition to expose the configuration UI
// in the dashboard. API authentication still applies to an unconfigured
// Chain Core.
func RunUnconfigured(ctx context.Context, db pg.DB, routableAddress string, opts ...RunOption) *API {
	a := &API{
		db:           db,
		accessTokens: &accesstoken.CredentialStore{DB: db},
		mux:          http.NewServeMux(),
		client:       rpcClient.NewURIClient(tendermintLAddr),
	}
	for _, opt := range opts {
		opt(a)
	}
/*	err := a.addAllowedMember(ctx, struct{ Addr string }{routableAddress})
	if err != nil {
		panic("failed to add self to member list: " + err.Error())
	}
	*/

	// Construct the complete http.Handler once.
	a.buildHandler()
	return a
}

// Run launches a new configured Chain Core. It will start goroutines
// for the various Core subsystems and enter leader election. It will not
// start listening for HTTP requests. To begin serving HTTP requests, use
// API.Handler to retrieve an http.Handler that can be used in a call to
// http.ListenAndServe.
//
// Either the GeneratorLocal or the GeneratorRemote RunOption is
// required.
func Run(
	ctx context.Context,
	db pg.DB,
	dbURL string,
	c *protocol.Chain,
	store *txdb.Store,
	routableAddress string,
	opts ...RunOption) (*API, error) {
	// Set up the pin store for block processing
	pinStore := pin.NewStore(db)
	err := pinStore.LoadAll(ctx)
	if err != nil {
		return nil, err
	}
	// Start listeners
	go pinStore.Listen(ctx, account.PinName, dbURL)
	go pinStore.Listen(ctx, account.ExpirePinName, dbURL)
	go pinStore.Listen(ctx, account.DeleteSpentsPinName, dbURL)
	go pinStore.Listen(ctx, asset.PinName, dbURL)

	assets := asset.NewRegistry(db, c, pinStore)
	accounts := account.NewManager(db, c, pinStore)
	indexer := query.NewIndexer(db, c, pinStore)

	a := &API{
		chain:        c,
		store:        store,
		pinStore:     pinStore,
		assets:       assets,
		accounts:     accounts,
		txFeeds:      &txfeed.Tracker{DB: db},
		indexer:      indexer,
		accessTokens: &accesstoken.CredentialStore{DB: db},
		db:           db,
		client:       rpcClient.NewURIClient(tendermintLAddr),
		mux:          http.NewServeMux(),
		addr:         routableAddress,
	}
	for _, opt := range opts {
		opt(a)
	}
	/*if a.remoteGenerator == nil && a.generator == nil {
		return nil, errors.New("no generator configured")
	}*/

	if a.indexTxs {
		go pinStore.Listen(ctx, query.TxPinName, dbURL)
		a.indexer.RegisterAnnotator(a.assets.AnnotateTxs)
		a.indexer.RegisterAnnotator(a.accounts.AnnotateTxs)
		a.assets.IndexAssets(a.indexer)
		a.accounts.IndexAccounts(a.indexer)
	}

	// Clean up expired UTXO reservations periodically.
	go accounts.ExpireReservations(ctx, expireReservationsPeriod)

	// GC old submitted txs periodically.
	go cleanUpSubmittedTxs(ctx, a.db)

	// When this cored becomes leader, run a.lead to perform
	// leader-only Core duties.
	//a.leader = leader.Run(ctx, db, routableAddress, a.lead)

	/*err = a.addAllowedMember(ctx, struct{ Addr string }{routableAddress})
	if err != nil {
		return nil, err
	}
	*/

	// Construct the complete http.Handler once.
	a.buildHandler()

	return a, nil
}
