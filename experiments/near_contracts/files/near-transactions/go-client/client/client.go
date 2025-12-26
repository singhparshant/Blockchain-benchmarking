package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/eteu-technologies/near-api-go/pkg/client/block"
	"github.com/eteu-technologies/near-api-go/pkg/jsonrpc"
	"github.com/eteu-technologies/near-api-go/pkg/types"
	"github.com/eteu-technologies/near-api-go/pkg/types/action"
	"github.com/eteu-technologies/near-api-go/pkg/types/hash"
	"github.com/eteu-technologies/near-api-go/pkg/types/key"
	"github.com/eteu-technologies/near-api-go/pkg/types/transaction"
)

var lastSent int64 = 0
var firstCall bool = true
var sendMutex sync.Mutex

type Client struct {
	RPCClient jsonrpc.Client
}

func NewClient(networkAddr string) (client Client, err error) {
	client.RPCClient, err = jsonrpc.NewClient(networkAddr)
	if err != nil {
		return
	}

	return
}

// https://docs.near.org/docs/api/rpc#send-transaction-async
func (c *Client) TransactionSend(ctx context.Context, from, to types.AccountID, actions []action.Action, sentAt *int64, spacingMicros int, txnOpts ...TransactionOpt) (res hash.CryptoHash, err error) {
	ctx2, blob, err := c.PrepareTransaction(ctx, from, to, actions, txnOpts...)
	if err != nil {
		return
	}
	return c.RPCTransactionSend(ctx2, blob, sentAt, spacingMicros)
}

// https://docs.near.org/docs/api/rpc#send-transaction-await
func (c *Client) TransactionSendAwait(ctx context.Context, from, to types.AccountID, actions []action.Action, sentAt *int64, txnOpts ...TransactionOpt) (res FinalExecutionOutcomeView, err error) {
	ctx2, blob, err := c.PrepareTransaction(ctx, from, to, actions, txnOpts...)
	if err != nil {
		return
	}
	return c.RPCTransactionSendAwait(ctx2, blob, sentAt)
}

func (c *Client) NetworkAddr() string {
	return c.RPCClient.URL
}

func (c *Client) doRPC(ctx context.Context, result interface{}, method string, block block.BlockCharacteristic, params interface{}) (res jsonrpc.Response, err error) {
	if block != nil {
		if mapv, ok := params.(map[string]interface{}); ok {
			block(mapv)
		}
	}

	res, err = c.RPCClient.CallRPC(ctx, method, params)
	if err != nil {
		return
	}

	// If JSON-RPC error happens, conveniently set it as err to avoid duplicating code
	// XXX: using plain assignment makes `err != nil` true for some reason
	if err := res.Error; err != nil {
		return res, err
	}

	if result != nil {
		if err = json.Unmarshal(res.Result, result); err != nil {
			return
		}
	}

	return
}

// https://docs.near.org/docs/api/rpc#send-transaction-async
// func (c *Client) RPCTransactionSend(ctx context.Context, signedTxnBase64 string, sentAt *int64) (resp hash.CryptoHash, err error) {
// 	*sentAt = time.Now().UnixNano()
// 	_, err = c.doRPC(ctx, &resp, "broadcast_tx_async", nil, []string{signedTxnBase64})

// 	return
// }

func (c *Client) RPCTransactionSend(ctx context.Context, signedTxnBase64 string, sentAt *int64, spacingMicros int) (resp hash.CryptoHash, err error) {
    if firstCall {
        lastSent = time.Now().UnixMicro()
        firstCall = false
    }
    startTime := time.Now().UnixMicro()
    // Calculate sinceLastSent based on lastSent from the previous invocation
    sinceLastSent := startTime - lastSent

    // RPC sending code
    *sentAt = time.Now().UnixMicro()
    _, err = c.doRPC(ctx, &resp, "broadcast_tx_async", nil, []string{signedTxnBase64})
    if err != nil {
		println("Error in RPC call. Exiting.")
        return;
    }

    endTime := time.Now().UnixMicro()
    elapsed := endTime - startTime

    // Converting durations to microseconds
    left := time.Duration(spacingMicros) * time.Microsecond
    right := time.Duration(sinceLastSent) * time.Microsecond
    diff := left - right

    adjustedSleep := diff - time.Duration(elapsed)*time.Microsecond // Adjust the sleep duration
    if adjustedSleep > 0 {
        time.Sleep(adjustedSleep)
    }

    // Update lastSent to reflect when this entire operation (including sleep) completed
    lastSent = time.Now().UnixMicro()

    return
}


func (c *Client) RPCTransactionSendAwait(ctx context.Context, signedTxnBase64 string, sentAt *int64) (resp FinalExecutionOutcomeView, err error) {
	*sentAt = time.Now().UnixNano()
	_, err = c.doRPC(ctx, &resp, "broadcast_tx_commit", nil, []string{signedTxnBase64})

	return
}

// https://docs.near.org/docs/api/rpc#block-details
func (c *Client) BlockDetails(ctx context.Context, block block.BlockCharacteristic) (resp BlockView, err error) {
	_, err = c.doRPC(ctx, &resp, "block", block, map[string]interface{}{})

	return
}

// https://docs.near.org/docs/api/rpc#transaction-status
func (c *Client) TransactionStatus(ctx context.Context, tx hash.CryptoHash, sender types.AccountID) (resp FinalExecutionOutcomeView, err error) {
	_, err = c.doRPC(ctx, &resp, "tx", nil, []string{
		tx.String(), sender,
	})

	return
}

// https://docs.near.org/docs/api/rpc#view-access-key
func (c *Client) AccessKeyView(ctx context.Context, accountID types.AccountID, publicKey key.Base58PublicKey, block block.BlockCharacteristic) (resp AccessKeyView, err error) {
	_, err = c.doRPC(ctx, &resp, "query", block, map[string]interface{}{
		"request_type": "view_access_key",
		"account_id":   accountID,
		"public_key":   publicKey,
	})

	if resp.Error != nil {
		err = fmt.Errorf("RPC returned an error: %w", errors.New(*resp.Error))
	}

	return
}

func WithBlockCharacteristic(block block.BlockCharacteristic) TransactionOpt {
	return func(ctx context.Context, txnCtx *transactionCtx) (err error) {
		client := ctx.Value(clientCtx).(*Client)

		var res BlockView
		if res, err = client.BlockDetails(ctx, block); err != nil {
			return
		}

		txnCtx.txn.BlockHash = res.Header.Hash
		return
	}

}

// WithLatestBlock is alias to `WithBlockCharacteristic(block.FinalityFinal())`
func WithLatestBlock() TransactionOpt {
	return WithBlockCharacteristic(block.FinalityFinal())
}

// WithKeyPair sets key pair to use sign this transaction with
func WithKeyPair(keyPair key.KeyPair) TransactionOpt {
	return func(_ context.Context, txnCtx *transactionCtx) (err error) {
		kp := keyPair
		txnCtx.keyPair = &kp
		return
	}
}

// Modified part of the API below

type transactionCtx struct {
	txn         transaction.Transaction
	keyPair     *key.KeyPair
	keyNonceSet bool
}

var nonceMutex = &sync.Mutex{}
var localNonce uint64
var nonceFetched bool

type TransactionOpt func(context.Context, *transactionCtx) error

func (c *Client) PrepareTransaction(ctx context.Context, from, to types.AccountID, actions []action.Action, txnOpts ...TransactionOpt) (ctx2 context.Context, blob string, err error) {
	ctx2 = context.WithValue(ctx, clientCtx, c)
	txn := transaction.Transaction{
		SignerID:   from,
		ReceiverID: to,
		Actions:    actions,
	}
	txnCtx := transactionCtx{
		txn:         txn,
		keyPair:     getKeyPair(ctx2),
		keyNonceSet: false,
	}

	for _, opt := range txnOpts {
		if err = opt(ctx2, &txnCtx); err != nil {
			return
		}
	}

	if txnCtx.keyPair == nil {
		err = errors.New("no keypair specified")
		return
	}

	txnCtx.txn.PublicKey = txnCtx.keyPair.PublicKey.ToPublicKey()
	// var mu sync.Mutex

	// Query the access key nonce, if not specified
	if !txnCtx.keyNonceSet {
		var err error
		txnCtx.txn.Nonce, err = getNextNonce(c, ctx2, txnCtx.txn.SignerID, txnCtx.keyPair.PublicKey)
		if err != nil {
			// Handle error
			log.Println(err)
			// return
		}
		txnCtx.keyNonceSet = true
	}

	blob, err = transaction.SignAndSerializeTransaction(*txnCtx.keyPair, txnCtx.txn)
	return
}

func getNextNonce(c *Client, ctx context.Context, signerID types.AccountID, publicKey key.Base58PublicKey) (uint64, error) {
	nonceMutex.Lock()
	defer nonceMutex.Unlock()

	// If nonce is not fetched yet, fetch from blockchain
	if !nonceFetched {
		var accessKey AccessKeyView
		var err error
		accessKey, err = c.AccessKeyView(ctx, signerID, publicKey, block.FinalityFinal())
		if err != nil {
			return 0, err
		}
		localNonce = accessKey.Nonce + 1
		nonceFetched = true
	} else {
		localNonce++
	}
	// log.Printf("nonce: %d", localNonce)
	return localNonce, nil
}
