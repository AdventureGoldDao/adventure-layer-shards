package ethapi

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

func (hb *HeatBeatAPI) sendHeatBeatTransaction(ctx context.Context, task *ContractTask, key *ecdsa.PrivateKey, data []byte) error {
	task.SendTxMutex.Lock()
	defer task.SendTxMutex.Unlock()

	fromAddr := crypto.PubkeyToAddress(key.PublicKey)

	gasLimit, err := hb.estimateGas(ctx, &fromAddr, &task.Address, data)
	if err != nil {
		return fmt.Errorf("gas estimation failed: %w", err)
	}

	gasPrice, err := hb.gasPrice(ctx)
	if err != nil {
		return fmt.Errorf("gas price fetch failed: %w", err)
	}

	nonce, err := hb.b.GetPoolNonce(ctx, fromAddr)
	if err != nil {
		return fmt.Errorf("nonce fetch failed: %w", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &task.Address,
		Value:    new(big.Int),
		Gas:      gasLimit,
		GasPrice: big.NewInt(gasPrice.ToInt().Int64() * defaultGasMultiplier),
		Data:     data,
	})

	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)
	if err != nil {
		return fmt.Errorf("tx signing failed: %w", err)
	}

	if _, err = SubmitTransaction(ctx, hb.b, signedTx); err != nil {
		return fmt.Errorf("tx submission failed: %w", err)
	}

	log.Info("Transaction sent", "hash", signedTx.Hash().Hex())
	return nil
}
