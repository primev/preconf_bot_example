// Package eth provides functionality for sending Ethereum transactions,
// including blob transactions with preconfirmation bids. This package
// is designed to work with public Ethereum nodes and a custom Titan
// endpoint for private transactions.
package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	bb "github.com/primev/preconf_blob_bidder/core/mevcommit"
)

// SelfETHTransfer sends an ETH transfer to the sender's own address. This function only works with
// public RPC endpoints and does not work with custom Titan endpoints.
//
// Parameters:
// - client: The Ethereum client instance.
// - authAcct: The authenticated account struct containing the address and private key.
// - value: The amount of ETH to transfer (in wei).
// - gasLimit: The maximum amount of gas to use for the transaction.
// - data: Optional data to include with the transaction.
//
// Returns:
// - The transaction hash as a string, or an error if the transaction fails.
func SelfETHTransfer(client *ethclient.Client, authAcct bb.AuthAcct, value *big.Int, offset uint64) (*types.Transaction, uint64, error) {
	// Get the account's nonce
	nonce, err := client.PendingNonceAt(context.Background(), authAcct.Address)
	if err != nil {
		return nil, 0, err
	}

	// Get the current base fee per gas from the latest block header
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, 0, err
	}
	baseFee := header.BaseFee

	
	blockNumber := header.Number.Uint64()

	// Set the max priority fee per gas to be 2 times the base fee
	maxPriorityFee := new(big.Int).Mul(baseFee, big.NewInt(2))

	// Set the max fee per gas to be 2 times the max priority fee
	maxFeePerGas := new(big.Int).Mul(maxPriorityFee, big.NewInt(2))

	// Get the chain ID (this does not work with the Titan RPC)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, 0, err
	}

	// Create a new EIP-1559 transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		To:        &authAcct.Address,
		Value:     value,
		Gas:       500_000,
		GasFeeCap: maxFeePerGas,
		GasTipCap: maxPriorityFee,
	})

	// Sign the transaction with the authenticated account's private key
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, authAcct.PrivateKey)
	if err != nil {
		log.Error("Failed to sign transaction", "error", err)
		return nil, 0, err
	}

	return signedTx, blockNumber + offset, nil

}