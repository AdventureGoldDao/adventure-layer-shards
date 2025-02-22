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

func (hb *HeartBeatAPI) sendHeartBeatTransaction(ctx context.Context, task *ContractTask, key *ecdsa.PrivateKey, data []byte) error {
	task.SendTxMutex.Lock()
	defer task.SendTxMutex.Unlock()

	fromAddr := crypto.PubkeyToAddress(key.PublicKey)

	gasLimit, err := hb.estimateGas(ctx, &fromAddr, &task.ContractAddress, data)
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
		To:       &task.ContractAddress,
		Value:    new(big.Int),
		Gas:      gasLimit,
		GasPrice: big.NewInt(gasPrice.ToInt().Int64() * defaultGasMultiplier),
		Data:     data,
	})

	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)
	if err != nil {
		return fmt.Errorf("tx signing failed: %w", err)
	}
	if err := checkTxFee(tx.GasPrice(), tx.Gas(), hb.b.RPCTxFeeCap()); err != nil {
		return fmt.Errorf("tx fee estimation failed: %w", err)
	}
	if err := hb.b.SendTx(ctx, signedTx); err != nil {
		return fmt.Errorf("tx sending failed: %w", err)
	}
	//go func() {
	//	var receipt *types.Receipt
	//	for {
	//		time.Sleep(2 * time.Second)
	//		receipts, err := hb.b.GetReceipts(ctx, signedTx.Hash())
	//		log.Info("GetReceipts", "receipts", receipts)
	//		if err != nil {
	//			log.Error("Failed to get receipts: %v", err)
	//		}
	//		if len(receipts) > 0 {
	//			receipt = receipts[0]
	//			break
	//		}
	//	}
	//
	//	actualGasFee := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), signedTx.GasPrice())
	//	log.Info("sendHeartBeatTransaction",
	//		"hash", signedTx.Hash().Hex(),
	//		"ContractAddress", task.ContractAddress.Hex(),
	//		"gas", signedTx.Gas(),
	//		"actualGasFee", actualGasFee.String(),
	//	)
	//}()
	log.Info("sendHeartBeatTransaction",
		"hash", signedTx.Hash().Hex(),
		"ContractAddress", task.ContractAddress.Hex(),
		"gas", signedTx.Gas(),
	)

	return nil
}
